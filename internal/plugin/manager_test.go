package plugin

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/perfect-panel/server/internal/config"
)

func TestManagerIsAllowed(t *testing.T) {
	m := &Manager{
		config: config.PluginConfig{
			AllowList: nil,
			BlockList: nil,
		},
	}

	// Empty lists → all allowed
	if !m.isAllowed("any-plugin") {
		t.Error("should allow when both lists empty")
	}
}

func TestManagerIsAllowedBlockList(t *testing.T) {
	m := &Manager{
		config: config.PluginConfig{
			BlockList: []string{"bad-plugin", "malware"},
		},
	}

	if !m.isAllowed("good-plugin") {
		t.Error("good-plugin should be allowed")
	}
	if m.isAllowed("bad-plugin") {
		t.Error("bad-plugin should be blocked")
	}
	if m.isAllowed("malware") {
		t.Error("malware should be blocked")
	}
}

func TestManagerIsAllowedAllowList(t *testing.T) {
	m := &Manager{
		config: config.PluginConfig{
			AllowList: []string{"approved-a", "approved-b"},
			BlockList: nil,
		},
	}

	if !m.isAllowed("approved-a") {
		t.Error("approved-a should be allowed")
	}
	if !m.isAllowed("approved-b") {
		t.Error("approved-b should be allowed")
	}
	if m.isAllowed("unknown-plugin") {
		t.Error("unknown-plugin should be blocked when allowlist is set")
	}
}

func TestManagerIsAllowedBothLists(t *testing.T) {
	// BlockList takes precedence over AllowList
	m := &Manager{
		config: config.PluginConfig{
			AllowList: []string{"plugin-a", "plugin-b"},
			BlockList: []string{"plugin-b"},
		},
	}

	if !m.isAllowed("plugin-a") {
		t.Error("plugin-a should be allowed")
	}
	if m.isAllowed("plugin-b") {
		t.Error("plugin-b should be blocked (blocklist overrides allowlist)")
	}
}

func TestValidatePluginName(t *testing.T) {
	valid := []string{"demo", "demo-plugin", "demo_plugin", "demo.plugin", "plugin1"}
	for _, name := range valid {
		if err := ValidatePluginName(name); err != nil {
			t.Fatalf("expected %q to be valid: %v", name, err)
		}
	}

	invalid := []string{"", "../demo", "demo/other", ".hidden", "demo..other", "Demo"}
	for _, name := range invalid {
		if err := ValidatePluginName(name); err == nil {
			t.Fatalf("expected %q to be invalid", name)
		}
	}
}

func TestManagerLoadPluginsCreatesMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "plugins")
	m := &Manager{
		config:  config.PluginConfig{Directory: dir},
		plugins: make(map[string]*PluginInstance),
	}
	m.loadPlugins(t.Context())

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected plugin directory to be created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected plugin path to be a directory")
	}
	if len(m.plugins) != 0 {
		t.Errorf("expected 0 plugins from empty dir, got %d", len(m.plugins))
	}
}

func TestManagerInstallPluginArchive(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(&HostEnv{
		Config: config.Config{
			Plugin: config.PluginConfig{Directory: dir, Enabled: true},
		},
	})

	body := makePluginArchive(t, map[string]string{
		"demo-plugin/plugin.yaml": `name: demo
version: 1.0
description: Demo plugin
author: PPanel
main: plugin.wasm
permissions:
  - logging
`,
		"demo-plugin/plugin.wasm": "not a real wasm module",
	})

	result, err := m.InstallPluginArchive(t.Context(), bytes.NewReader(body), PluginInstallOptions{})
	requireNoError(t, err)
	if result.Name != "demo" {
		t.Fatalf("installed plugin name = %q, want demo", result.Name)
	}
	if result.Status != StatusUnloaded {
		t.Fatalf("installed plugin status = %q, want unloaded", result.Status)
	}
	if _, err := os.Stat(filepath.Join(dir, "demo", "plugin.yaml")); err != nil {
		t.Fatalf("expected plugin manifest to be installed: %v", err)
	}
	if result.Validation.Valid {
		t.Fatalf("dummy wasm should not compile")
	}
}

func TestManagerInstallPluginArchiveRejectsDuplicateUnlessReplace(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(&HostEnv{
		Config: config.Config{
			Plugin: config.PluginConfig{Directory: dir, Enabled: true},
		},
	})
	body := makePluginArchive(t, map[string]string{
		"plugin.yaml": `name: demo
version: 1.0
main: plugin.wasm
permissions: []
`,
		"plugin.wasm": "wasm",
	})

	_, err := m.InstallPluginArchive(t.Context(), bytes.NewReader(body), PluginInstallOptions{})
	requireNoError(t, err)
	if _, err := m.InstallPluginArchive(t.Context(), bytes.NewReader(body), PluginInstallOptions{}); err == nil {
		t.Fatalf("expected duplicate install to fail")
	}
	result, err := m.InstallPluginArchive(t.Context(), bytes.NewReader(body), PluginInstallOptions{Replace: true})
	requireNoError(t, err)
	if !result.Replaced {
		t.Fatalf("expected replace result")
	}
}

func TestManagerInstallPluginArchiveRejectsUnsafeEntry(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(&HostEnv{
		Config: config.Config{
			Plugin: config.PluginConfig{Directory: dir, Enabled: true},
		},
	})
	body := makePluginArchive(t, map[string]string{
		"../escape.txt": "bad",
	})

	if _, err := m.InstallPluginArchive(t.Context(), bytes.NewReader(body), PluginInstallOptions{}); err == nil {
		t.Fatalf("expected unsafe zip entry to fail")
	}
}

func TestManagerListInstalledPluginsIncludesUnloaded(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "demo")
	requireNoError(t, os.Mkdir(pluginDir, 0755))
	requireNoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(`name: demo
version: 1.0
description: Demo plugin
author: PPanel
main: plugin.wasm
permissions:
  - logging
`), 0644))

	m := &Manager{
		config:  config.PluginConfig{Directory: dir},
		plugins: make(map[string]*PluginInstance),
	}

	plugins := m.ListInstalledPlugins()
	if len(plugins) != 1 {
		t.Fatalf("expected 1 installed plugin, got %d", len(plugins))
	}
	if plugins[0].Name != "demo" {
		t.Fatalf("expected plugin name demo, got %q", plugins[0].Name)
	}
	if plugins[0].Status != StatusUnloaded {
		t.Fatalf("expected unloaded status, got %q", plugins[0].Status)
	}
}

func TestManagerRuntimeIntrospection(t *testing.T) {
	p := &PluginInstance{
		Name:     "demo",
		Status:   StatusRunning,
		Routes:   []RouteRegistration{{PluginName: "demo", Method: "GET", Path: "/hello", Handler: "handle_hello"}},
		Pool:     NewInstancePool(1),
		AsyncSem: make(chan struct{}, 2),
	}
	p.AsyncSem <- struct{}{}
	m := &Manager{
		plugins: map[string]*PluginInstance{"demo": p},
		routes: map[string]RouteRegistration{
			routeKey("demo", "GET", "/hello"): {PluginName: "demo", Method: "GET", Path: "/hello", Handler: "handle_hello"},
		},
		middleware: map[string]MiddlewareRegistration{
			middlewareKey("demo", "guard"): {PluginName: "demo", Name: "guard", Handler: "mw_guard"},
		},
		eventBus: NewEventBus(nil),
	}
	m.eventBus.Subscribe(EventSubscription{PluginName: "demo", Event: "demo.ping", Handler: "on_demo_ping"})

	health, ok := m.GetPluginHealth("demo")
	if !ok {
		t.Fatal("expected health for demo")
	}
	if !health.Ready || health.PoolSize != 1 || health.AsyncInFlight != 1 || health.AsyncLimit != 2 || health.RegisteredRoute != 1 {
		t.Fatalf("unexpected health: %#v", health)
	}
	if routes := m.ListPluginRoutes("demo"); len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if middleware := m.ListPluginMiddleware("demo"); len(middleware) != 1 {
		t.Fatalf("expected 1 middleware, got %d", len(middleware))
	}
	if events := m.ListPluginEvents("demo"); len(events) != 1 {
		t.Fatalf("expected 1 event subscription, got %d", len(events))
	}
}

func TestManagerValidateInstalledPluginRejectsUnsafeMain(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "demo")
	requireNoError(t, os.Mkdir(pluginDir, 0755))
	requireNoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(`name: demo
version: 1.0
main: ../escape.wasm
permissions: []
`), 0644))

	m := NewManager(&HostEnv{
		Config: config.Config{
			Plugin: config.PluginConfig{Directory: dir, Enabled: true},
		},
	})

	validation := m.ValidateInstalledPlugin("demo")
	if validation.Valid {
		t.Fatalf("expected validation to fail for unsafe main path")
	}
	if validation.Error == "" {
		t.Fatalf("expected validation error")
	}
}

func TestManagerReloadAllPluginsClearsRuntimeRegistrations(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "no-wasm")
	requireNoError(t, os.Mkdir(pluginDir, 0755))
	requireNoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(`name: no-wasm
version: 1.0
main: plugin.wasm
permissions: []
`), 0644))

	m := NewManager(&HostEnv{
		Config: config.Config{
			Plugin: config.PluginConfig{Directory: dir, Enabled: true},
		},
	})
	m.plugins["old"] = &PluginInstance{Name: "old", Status: StatusRunning}
	m.routes[routeKey("old", "GET", "/old")] = RouteRegistration{
		PluginName: "old",
		Method:     "GET",
		Path:       "/old",
		Handler:    "handle_old",
	}

	plugins := m.ReloadAllPlugins(t.Context())
	if _, ok := m.FindRoute("old", "GET", "/old"); ok {
		t.Fatalf("expected old route registration to be cleared")
	}
	if len(plugins) != 1 || plugins[0].Name != "no-wasm" {
		t.Fatalf("unexpected reloaded plugin list: %#v", plugins)
	}
	if p := m.GetPlugin("no-wasm"); p == nil || p.Status != StatusError {
		t.Fatalf("expected no-wasm plugin to be loaded with error status")
	}
}

func TestManagerLoadPluginsInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "broken-plugin")
	os.Mkdir(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte("{{{bad"), 0644)

	m := &Manager{
		config:  config.PluginConfig{Directory: dir},
		plugins: make(map[string]*PluginInstance),
	}
	m.loadPlugins(t.Context())

	if len(m.plugins) != 0 {
		t.Errorf("expected 0 plugins from broken manifest, got %d", len(m.plugins))
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestManagerLoadPluginsValidManifestNoWasm(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "no-wasm")
	os.Mkdir(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(`name: no-wasm
version: 1.0
main: plugin.wasm
permissions: []
`), 0644)
	// No .wasm file → compile error → plugin stored with error status

	m := &Manager{
		config:  config.PluginConfig{Directory: dir},
		plugins: make(map[string]*PluginInstance),
	}
	m.loadPlugins(t.Context())

	p := m.GetPlugin("no-wasm")
	if p == nil {
		t.Fatal("plugin should exist even with compile error")
	}
	if p.Status != StatusError {
		t.Errorf("expected StatusError, got %q", p.Status)
	}
	if p.Error == "" {
		t.Error("expected error message for missing wasm file")
	}
}

func TestDefaultPoolSizeKeepsLifecycleSingleShot(t *testing.T) {
	if defaultPoolSize != 1 {
		t.Fatalf("defaultPoolSize = %d, want 1 to avoid duplicated init/start side effects", defaultPoolSize)
	}
}

func TestManagerGetConfigValue(t *testing.T) {
	m := &Manager{
		hostEnv: &HostEnv{
			Config: config.Config{
				Host:  "0.0.0.0",
				Port:  8080,
				Debug: true,
				Site: config.SiteConfig{
					SiteName: "TestPanel",
					Host:     "panel.example.com",
				},
				Currency: config.Currency{
					Unit:   "USD",
					Symbol: "$",
				},
			},
		},
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"Host", "0.0.0.0"},
		{"Port", "8080"},
		{"Debug", "true"},
		{"Site.SiteName", "TestPanel"},
		{"Site.Host", "panel.example.com"},
		{"Currency.Unit", "USD"},
		{"Currency.Symbol", "$"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		if got := m.GetConfigValue(tt.key); got != tt.expected {
			t.Errorf("GetConfigValue(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestManagerRegisterRouteDedupAndFind(t *testing.T) {
	p := &PluginInstance{Name: "demo", Manifest: &PluginManifest{Name: "demo"}}
	m := &Manager{
		plugins: map[string]*PluginInstance{"demo": p},
		routes:  make(map[string]RouteRegistration),
	}

	route := RouteRegistration{PluginName: "demo", Method: "get", Path: "hello", Handler: "handle_hello"}
	if err := m.RegisterRoute(route); err != nil {
		t.Fatalf("RegisterRoute error: %v", err)
	}
	if err := m.RegisterRoute(route); err != nil {
		t.Fatalf("RegisterRoute duplicate error: %v", err)
	}

	routes := m.PendingRoutes()
	if len(routes) != 1 {
		t.Fatalf("routes len = %d, want 1", len(routes))
	}
	if routes[0].Method != "GET" || routes[0].Path != "/hello" {
		t.Fatalf("route normalized to %s %s, want GET /hello", routes[0].Method, routes[0].Path)
	}
	found, ok := m.FindRoute("demo", "GET", "/hello")
	if !ok {
		t.Fatal("FindRoute did not find registered route")
	}
	if found.Handler != "handle_hello" {
		t.Fatalf("handler = %q, want handle_hello", found.Handler)
	}
	if got := p.ToInfo().Routes; len(got) != 1 || got[0] != "GET /hello" {
		t.Fatalf("plugin info routes = %#v, want [GET /hello]", got)
	}
}

func TestManagerRegisterMiddlewareDedupAndFind(t *testing.T) {
	m := &Manager{middleware: make(map[string]MiddlewareRegistration)}
	mw := MiddlewareRegistration{PluginName: "demo", Name: "authz", Handler: "mw_authz"}
	if err := m.RegisterMiddleware(mw); err != nil {
		t.Fatalf("RegisterMiddleware error: %v", err)
	}
	if err := m.RegisterMiddleware(mw); err != nil {
		t.Fatalf("RegisterMiddleware duplicate error: %v", err)
	}
	if got := m.PendingMiddleware(); len(got) != 1 {
		t.Fatalf("middleware len = %d, want 1", len(got))
	}
	found, ok := m.FindMiddleware("demo", "authz")
	if !ok {
		t.Fatal("FindMiddleware did not find registered middleware")
	}
	if found.Handler != "mw_authz" {
		t.Fatalf("handler = %q, want mw_authz", found.Handler)
	}
}

func TestResolveAsyncTaskNotFoundIsDone(t *testing.T) {
	m := &Manager{asyncTasks: make(map[uint64]*asyncTask)}
	result, errStr, done := m.resolveAsyncTask("demo", 42)
	if !done {
		t.Fatal("missing task should be reported as done with an error")
	}
	if result != nil || errStr == "" {
		t.Fatalf("result = %#v, error = %q; want nil result and non-empty error", result, errStr)
	}
}

func TestSanitizeEventName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"order.paid", "order_paid"},
		{"user.created", "user_created"},
		{"simple", "simple"},
		{"with-hyphen", "with_hyphen"},
		{"", ""},
		{"UPPER.case", "UPPER_case"},
	}

	for _, tt := range tests {
		if got := sanitizeEventName(tt.input); got != tt.expected {
			t.Errorf("sanitizeEventName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func makePluginArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		requireNoError(t, err)
		_, err = w.Write([]byte(body))
		requireNoError(t, err)
	}
	requireNoError(t, zw.Close())
	return buf.Bytes()
}
