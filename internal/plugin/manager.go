package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	pluginv1 "github.com/perfect-panel/server/api/plugin/v1"
	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/robfig/cron/v3"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Manager 插件管理器，实现 service.Service 接口
type Manager struct {
	hostEnv *HostEnv
	config  config.PluginConfig
	plugins map[string]*PluginInstance
	mu      sync.RWMutex

	// pendingRoutes 保留历史 API 名称，实际作为动态路由注册表的快照
	pendingRoutes []RouteRegistration
	routes        map[string]RouteRegistration

	// pendingMiddleware 保留历史 API 名称，实际作为动态中间件注册表的快照
	pendingMiddleware []MiddlewareRegistration
	middleware        map[string]MiddlewareRegistration

	// 事件总线
	eventBus *EventBus

	// cronScheduler 定时任务调度器
	cronScheduler *cron.Cron

	// scheduledTasks 记录插件注册的定时任务（用于停止时清理）
	scheduledTasks []scheduledTask

	// asyncTaskMu 和以下字段管理异步任务（goroutine 池）
	asyncTaskMu  sync.Mutex
	asyncTasks   map[uint64]*asyncTask
	asyncTaskSeq uint64

	// readyCh 当插件加载完成时关闭
	readyCh chan struct{}
}

type asyncTask struct {
	id         uint64
	pluginName string
	opType     string
	params     []byte
	result     []byte
	err        string
	done       chan struct{}
	notifyCh   chan uint64
	release    func()
}

type scheduledTask struct {
	pluginName string
	taskName   string
	cronExpr   string
	handler    string
	entryID    cron.EntryID
}

// NewManager 创建插件管理器
func NewManager(hostEnv *HostEnv) *Manager {
	cfg := hostEnv.Config.Plugin
	m := &Manager{
		hostEnv:       hostEnv,
		config:        cfg,
		plugins:       make(map[string]*PluginInstance),
		routes:        make(map[string]RouteRegistration),
		middleware:    make(map[string]MiddlewareRegistration),
		readyCh:       make(chan struct{}),
		cronScheduler: cron.New(cron.WithParser(cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow))),
		asyncTasks:    make(map[uint64]*asyncTask),
	}
	m.eventBus = NewEventBus(m.onPluginEvent)
	return m
}

// Start 启动插件管理器（实现 service.Service）
func (m *Manager) Start() {
	ctx := context.Background()

	if !m.config.Enabled {
		close(m.readyCh)
		return
	}

	m.cronScheduler.Start()
	m.loadPlugins(ctx)
	close(m.readyCh)
}

// Stop 停止插件管理器（实现 service.Service）
func (m *Manager) Stop() {
	ctx := context.Background()
	m.cronScheduler.Stop()

	m.mu.Lock()
	plugins := make([]*PluginInstance, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	m.plugins = make(map[string]*PluginInstance)
	m.pendingRoutes = nil
	m.routes = make(map[string]RouteRegistration)
	m.pendingMiddleware = nil
	m.middleware = make(map[string]MiddlewareRegistration)
	m.scheduledTasks = nil
	m.mu.Unlock()

	for _, p := range plugins {
		m.stopPlugin(ctx, p)
	}
}

func (m *Manager) stopPlugin(ctx context.Context, p *PluginInstance) {
	if p == nil {
		return
	}
	if p.GetStatus() == StatusRunning && p.Pool != nil {
		stopCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if mod, err := p.Pool.Get(stopCtx); err == nil {
			if err := CallGuestStop(stopCtx, mod); err != nil {
				logger.Errorf("plugin %q stop error: %v", p.Name, err)
			}
			p.Pool.Put(mod)
		}
		cancel()
		p.Pool.Close(ctx)
	}
	if p.Runtime != nil {
		if err := p.Runtime.Close(ctx); err != nil {
			logger.Errorf("plugin %q runtime close error: %v", p.Name, err)
		}
		p.Runtime = nil
	}
	p.SetStatus(StatusStopped)
}

// PendingRoutes 返回待注册的路由列表
func (m *Manager) PendingRoutes() []RouteRegistration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]RouteRegistration, len(m.pendingRoutes))
	copy(result, m.pendingRoutes)
	return result
}

// FindRoute 查找当前已注册的插件路由。用于固定 catch-all HTTP 路由动态分发。
func (m *Manager) FindRoute(pluginName, method, path string) (RouteRegistration, bool) {
	method = normalizeMethod(method)
	path = normalizePath(path)

	m.mu.RLock()
	defer m.mu.RUnlock()
	if route, ok := m.routes[routeKey(pluginName, method, path)]; ok {
		return route, true
	}
	if route, ok := m.routes[routeKey(pluginName, "ANY", path)]; ok {
		return route, true
	}
	return RouteRegistration{}, false
}

// HasPermission 检查插件是否有特定权限
func (m *Manager) HasPermission(pluginName string, perm Permission) bool {
	p := m.GetPlugin(pluginName)
	if p == nil || p.Manifest == nil {
		return false
	}
	for _, prm := range p.Manifest.Permissions {
		if Permission(prm) == perm {
			return true
		}
	}
	return false
}

// ValidatePluginName 校验管理 API 传入的插件名称，避免路径穿越。
func ValidatePluginName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if len(name) > 64 {
		return fmt.Errorf("plugin name is too long")
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			continue
		}
		return fmt.Errorf("plugin name contains invalid character %q", c)
	}
	if name[0] == '.' || strings.Contains(name, "..") {
		return fmt.Errorf("plugin name must not contain path traversal")
	}
	return nil
}

func normalizePluginName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if err := ValidatePluginName(name); err != nil {
		return "", err
	}
	return name, nil
}

func (m *Manager) pluginDirectory() string {
	pluginDir := m.config.Directory
	if pluginDir == "" {
		pluginDir = "plugins"
	}
	return pluginDir
}

// PluginDirectory 返回当前插件目录。
func (m *Manager) PluginDirectory() string {
	return m.pluginDirectory()
}

func (m *Manager) pluginPath(name string) (string, error) {
	name, err := normalizePluginName(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(m.pluginDirectory(), name), nil
}

// GetPlugin 根据名称获取插件实例
func (m *Manager) GetPlugin(name string) *PluginInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[name]
}

// ListPlugins 返回所有已加载插件信息
func (m *Manager) ListPlugins() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]PluginInfo, 0, len(m.plugins))
	for _, p := range m.plugins {
		result = append(result, p.ToInfo())
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListInstalledPlugins 扫描插件目录，并合并当前运行时状态。
func (m *Manager) ListInstalledPlugins() []PluginInfo {
	m.mu.RLock()
	loaded := make(map[string]PluginInfo, len(m.plugins))
	for name, p := range m.plugins {
		loaded[name] = p.ToInfo()
	}
	m.mu.RUnlock()

	result := make([]PluginInfo, 0, len(loaded))
	seen := make(map[string]bool, len(loaded))

	entries, err := os.ReadDir(m.pluginDirectory())
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if info, ok := loaded[name]; ok {
				result = append(result, info)
				seen[name] = true
				continue
			}

			manifest, err := ParseManifest(filepath.Join(m.pluginDirectory(), name))
			if err != nil {
				result = append(result, PluginInfo{
					Name:   name,
					Status: StatusError,
					Error:  err.Error(),
				})
				seen[name] = true
				continue
			}

			perms := make([]string, len(manifest.Permissions))
			copy(perms, manifest.Permissions)
			result = append(result, PluginInfo{
				Name:        manifest.Name,
				Version:     manifest.Version,
				Description: manifest.Description,
				Author:      manifest.Author,
				Status:      StatusUnloaded,
				Permissions: perms,
				Routes:      []string{},
			})
			seen[manifest.Name] = true
		}
	}

	for name, info := range loaded {
		if !seen[name] {
			result = append(result, info)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// GetInstalledPluginInfo 返回运行中或已安装插件的信息。
func (m *Manager) GetInstalledPluginInfo(name string) (PluginInfo, bool) {
	name, err := normalizePluginName(name)
	if err != nil {
		return PluginInfo{Error: err.Error()}, false
	}
	if p := m.GetPlugin(name); p != nil {
		return p.ToInfo(), true
	}

	manifest, err := ParseManifest(filepath.Join(m.pluginDirectory(), name))
	if err != nil {
		return PluginInfo{}, false
	}
	perms := make([]string, len(manifest.Permissions))
	copy(perms, manifest.Permissions)
	return PluginInfo{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Status:      StatusUnloaded,
		Permissions: perms,
		Routes:      []string{},
	}, true
}

// GetInstalledManifest 返回运行中或已安装插件的 manifest。
func (m *Manager) GetInstalledManifest(name string) (*PluginManifest, error) {
	name, err := normalizePluginName(name)
	if err != nil {
		return nil, err
	}
	if p := m.GetPlugin(name); p != nil && p.Manifest != nil {
		return p.Manifest, nil
	}
	pluginPath, err := m.pluginPath(name)
	if err != nil {
		return nil, err
	}
	return ParseManifest(pluginPath)
}

// ListPluginRoutes 返回指定插件当前注册的路由。
func (m *Manager) ListPluginRoutes(name string) []RouteRegistration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]RouteRegistration, 0)
	for _, route := range m.routes {
		if route.PluginName == name {
			result = append(result, route)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			return result[i].Method < result[j].Method
		}
		return result[i].Path < result[j].Path
	})
	return result
}

// ListPluginMiddleware 返回指定插件当前注册的 WASM 中间件。
func (m *Manager) ListPluginMiddleware(name string) []MiddlewareRegistration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]MiddlewareRegistration, 0)
	for _, mw := range m.middleware {
		if mw.PluginName == name {
			result = append(result, mw)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListPluginEvents 返回指定插件当前订阅的事件。
func (m *Manager) ListPluginEvents(name string) []EventSubscription {
	result := m.eventBus.Subscriptions(name)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Event == result[j].Event {
			return result[i].Handler < result[j].Handler
		}
		return result[i].Event < result[j].Event
	})
	return result
}

// GetPluginHealth 返回指定插件运行时健康状态。
func (m *Manager) GetPluginHealth(name string) (PluginHealth, bool) {
	name, err := normalizePluginName(name)
	if err != nil {
		return PluginHealth{Name: strings.TrimSpace(name), Status: StatusError, Error: err.Error()}, false
	}
	p := m.GetPlugin(name)
	if p == nil {
		return PluginHealth{Name: name, Status: StatusUnloaded}, true
	}
	p.mu.RLock()
	status := p.Status
	errMsg := p.Error
	poolSize := 0
	if p.Pool != nil {
		poolSize = p.Pool.Size()
	}
	asyncInFlight := 0
	asyncLimit := 0
	if p.AsyncSem != nil {
		asyncInFlight = len(p.AsyncSem)
		asyncLimit = cap(p.AsyncSem)
	}
	routeCount := len(p.Routes)
	p.mu.RUnlock()

	return PluginHealth{
		Name:            name,
		Status:          status,
		Ready:           status == StatusRunning && p.Pool != nil,
		PoolSize:        poolSize,
		AsyncInFlight:   asyncInFlight,
		AsyncLimit:      asyncLimit,
		RegisteredRoute: routeCount,
		Error:           errMsg,
	}, true
}

// ValidateInstalledPlugin 校验插件目录、manifest、权限和 WASM 文件可访问性。
func (m *Manager) ValidateInstalledPlugin(name string) PluginValidation {
	name = strings.TrimSpace(name)
	validation := PluginValidation{
		Name:   name,
		Checks: make([]PluginValidationCheck, 0, 4),
	}

	addCheck := func(name string, ok bool, message string) {
		validation.Checks = append(validation.Checks, PluginValidationCheck{
			Name:    name,
			OK:      ok,
			Message: message,
		})
		if !ok && validation.Error == "" {
			validation.Error = message
		}
	}

	if err := ValidatePluginName(name); err != nil {
		addCheck("plugin_name", false, err.Error())
		return validation
	}
	addCheck("plugin_name", true, "plugin name is valid")

	pluginPath, err := m.pluginPath(name)
	if err != nil {
		addCheck("plugin_path", false, err.Error())
		return validation
	}
	info, err := os.Stat(pluginPath)
	if err != nil {
		addCheck("plugin_path", false, err.Error())
		return validation
	}
	if !info.IsDir() {
		addCheck("plugin_path", false, "plugin path is not a directory")
		return validation
	}
	addCheck("plugin_path", true, "plugin directory exists")

	manifest, err := ParseManifest(pluginPath)
	if err != nil {
		addCheck("manifest", false, err.Error())
		return validation
	}
	validation.Manifest = manifest
	if manifest.Name != name {
		addCheck("manifest", false, fmt.Sprintf("manifest name %q does not match directory name %q", manifest.Name, name))
		return validation
	}
	addCheck("manifest", true, "manifest is valid")

	if !m.isAllowed(manifest.Name) {
		addCheck("allowlist", false, "plugin is blocked by allowlist/blocklist")
		return validation
	}
	addCheck("allowlist", true, "plugin is allowed by policy")

	wasmPath, err := ResolvePluginFile(pluginPath, manifest.Main)
	if err != nil {
		addCheck("wasm", false, err.Error())
		return validation
	}
	wasmInfo, err := os.Stat(wasmPath)
	if err != nil {
		addCheck("wasm", false, err.Error())
		return validation
	}
	if wasmInfo.IsDir() {
		addCheck("wasm", false, "wasm path is a directory")
		return validation
	}
	addCheck("wasm", true, "wasm file exists")

	validateCtx, cancel := context.WithTimeout(context.Background(), m.RequestTimeout())
	defer cancel()
	rt, err := m.newPluginRuntime(validateCtx)
	if err != nil {
		addCheck("compile", false, err.Error())
		return validation
	}
	defer rt.Close(validateCtx)
	compiled, err := CompilePlugin(validateCtx, rt, pluginPath, manifest.Main)
	if err != nil {
		addCheck("compile", false, err.Error())
		return validation
	}
	compiled.Close(validateCtx)
	addCheck("compile", true, "wasm module compiles")

	validation.Valid = true
	return validation
}

// ReloadAllPlugins 停止当前运行时状态并重新扫描插件目录。
func (m *Manager) ReloadAllPlugins(ctx context.Context) []PluginInfo {
	m.mu.Lock()
	plugins := make([]*PluginInstance, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, p)
	}
	for _, st := range m.scheduledTasks {
		m.cronScheduler.Remove(st.entryID)
	}
	m.plugins = make(map[string]*PluginInstance)
	m.pendingRoutes = nil
	m.routes = make(map[string]RouteRegistration)
	m.pendingMiddleware = nil
	m.middleware = make(map[string]MiddlewareRegistration)
	m.scheduledTasks = nil
	m.mu.Unlock()

	m.asyncTaskMu.Lock()
	m.asyncTasks = make(map[uint64]*asyncTask)
	m.asyncTaskMu.Unlock()

	if m.eventBus != nil {
		for _, p := range plugins {
			m.eventBus.Unsubscribe(p.Name)
		}
	}
	for _, p := range plugins {
		m.stopPlugin(ctx, p)
	}
	m.loadPlugins(ctx)
	return m.ListInstalledPlugins()
}

// CallPlugin 调用插件的导出函数（用于路由处理）。
// 内部持有 WASM 互斥锁，阻塞型宿主函数安全。
func (m *Manager) CallPlugin(ctx context.Context, pluginName, exportName string, req *pluginv1.HandleRequest) (*pluginv1.HandleResponse, error) {
	p := m.GetPlugin(pluginName)
	if p == nil || p.Pool == nil {
		return nil, fmt.Errorf("plugin %q not found or not ready", pluginName)
	}
	return callGuestFromPool(ctx, p, exportName, req,
		func() *pluginv1.HandleResponse { return &pluginv1.HandleResponse{} },
	)
}

// CallPluginMiddleware 调用插件中间件
func (m *Manager) CallPluginMiddleware(ctx context.Context, pluginName, exportName string, req *pluginv1.HandleRequest) (*pluginv1.MiddlewareResponse, error) {
	p := m.GetPlugin(pluginName)
	if p == nil || p.Pool == nil {
		return nil, fmt.Errorf("plugin %q not found or not ready", pluginName)
	}
	return callGuestFromPool(ctx, p, exportName, req,
		func() *pluginv1.MiddlewareResponse { return &pluginv1.MiddlewareResponse{} },
	)
}

// callGuestFromPool 从实例池获取 WASM 模块并调用 guest 导出函数
func callGuestFromPool[T, R proto.Message](
	ctx context.Context,
	p *PluginInstance,
	exportName string,
	req T,
	newR func() R,
) (R, error) {
	mod, err := p.Pool.Get(ctx)
	if err != nil {
		var zero R
		return zero, err
	}
	defer p.Pool.Put(mod)
	return CallGuestFn(ctx, mod, exportName, req, newR)
}

// buildInitRequest 为 guest init 调用构造 InitRequest
func (m *Manager) buildInitRequest(manifest *PluginManifest) *pluginv1.InitRequest {
	// 将插件 config 转换为 protobuf Struct
	cfgStruct, _ := structpb.NewStruct(manifest.Config)

	hostCfg := m.hostEnv.Config
	req := &pluginv1.InitRequest{
		PluginName: manifest.Name,
		Config:     cfgStruct,
		HostConfig: &pluginv1.HostConfig{
			SiteName:       hostCfg.Site.SiteName,
			SiteHost:       hostCfg.Site.Host,
			CurrencyUnit:   hostCfg.Currency.Unit,
			CurrencySymbol: hostCfg.Currency.Symbol,
			Debug:          hostCfg.Debug,
		},
	}
	return req
}

// WaitReady 阻塞直到插件管理器加载完成
func (m *Manager) WaitReady(ctx context.Context) error {
	select {
	case <-m.readyCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RegisterRoute 注册插件的 HTTP 路由（供 host_register_route 调用）
func (m *Manager) RegisterRoute(route RouteRegistration) error {
	route.PluginName = strings.TrimSpace(route.PluginName)
	route.Method = normalizeMethod(route.Method)
	route.Path = normalizePath(route.Path)
	route.Handler = strings.TrimSpace(route.Handler)
	if route.PluginName == "" {
		return fmt.Errorf("plugin name is required")
	}
	if route.Method == "" {
		return fmt.Errorf("method is required")
	}
	if route.Path == "" {
		return fmt.Errorf("path is required")
	}
	if route.Handler == "" {
		return fmt.Errorf("handler is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.routes == nil {
		m.routes = make(map[string]RouteRegistration)
	}
	key := routeKey(route.PluginName, route.Method, route.Path)
	if _, exists := m.routes[key]; exists {
		return nil
	}
	m.routes[key] = route
	m.pendingRoutes = append(m.pendingRoutes, route)

	if p := m.plugins[route.PluginName]; p != nil {
		p.AddRoute(route)
	}
	return nil
}

// RegisterMiddleware 注册插件中间件（供 host_register_middleware 调用）
func (m *Manager) RegisterMiddleware(mw MiddlewareRegistration) error {
	mw.PluginName = strings.TrimSpace(mw.PluginName)
	mw.Name = strings.TrimSpace(mw.Name)
	mw.Handler = strings.TrimSpace(mw.Handler)
	if mw.PluginName == "" {
		return fmt.Errorf("plugin name is required")
	}
	if mw.Name == "" {
		return fmt.Errorf("middleware name is required")
	}
	if mw.Handler == "" {
		return fmt.Errorf("middleware handler is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.middleware == nil {
		m.middleware = make(map[string]MiddlewareRegistration)
	}
	key := middlewareKey(mw.PluginName, mw.Name)
	if _, exists := m.middleware[key]; exists {
		return nil
	}
	m.middleware[key] = mw
	m.pendingMiddleware = append(m.pendingMiddleware, mw)
	return nil
}

// PendingMiddleware 返回待注册的中间件列表
func (m *Manager) PendingMiddleware() []MiddlewareRegistration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]MiddlewareRegistration, len(m.pendingMiddleware))
	copy(result, m.pendingMiddleware)
	return result
}

// FindMiddleware 查找插件中间件。优先使用同插件注册的中间件，兼容旧的全局名称引用。
func (m *Manager) FindMiddleware(pluginName, name string) (MiddlewareRegistration, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if mw, ok := m.middleware[middlewareKey(pluginName, name)]; ok {
		return mw, true
	}
	for _, mw := range m.pendingMiddleware {
		if mw.Name == name {
			return mw, true
		}
	}
	return MiddlewareRegistration{}, false
}

// EventBus 返回事件总线
func (m *Manager) EventBus() *EventBus {
	return m.eventBus
}

// GetRedis 返回 Redis 客户端
func (m *Manager) GetRedis() RedisClient {
	return m.hostEnv.Redis
}

// GetStore 返回数据库 Store
func (m *Manager) GetStore() StoreClient {
	return m.hostEnv.Store
}

// GetConfigValue 通过 dot-notation key 读取配置值
func (m *Manager) GetConfigValue(key string) string {
	cfg := m.hostEnv.Config
	switch key {
	case "Site.SiteName":
		return cfg.Site.SiteName
	case "Site.SiteDesc":
		return cfg.Site.SiteDesc
	case "Site.SiteLogo":
		return cfg.Site.SiteLogo
	case "Site.Host":
		return cfg.Site.Host
	case "Currency.Unit":
		return cfg.Currency.Unit
	case "Currency.Symbol":
		return cfg.Currency.Symbol
	case "Debug":
		if cfg.Debug {
			return "true"
		}
		return "false"
	case "Host":
		return cfg.Host
	case "Port":
		return fmt.Sprintf("%d", cfg.Port)
	default:
		return ""
	}
}

// ScheduleTask 注册定时任务
func (m *Manager) ScheduleTask(pluginName, taskName, cronExpr, handler string) error {
	pluginName = strings.TrimSpace(pluginName)
	taskName = strings.TrimSpace(taskName)
	cronExpr = strings.TrimSpace(cronExpr)
	handler = strings.TrimSpace(handler)
	if pluginName == "" || taskName == "" || cronExpr == "" || handler == "" {
		return fmt.Errorf("pluginName, taskName, cron and handler are required")
	}

	m.mu.RLock()
	for _, st := range m.scheduledTasks {
		if st.pluginName == pluginName && st.taskName == taskName {
			m.mu.RUnlock()
			return nil
		}
	}
	m.mu.RUnlock()

	entryID, err := m.cronScheduler.AddFunc(cronExpr, func() {
		p := m.GetPlugin(pluginName)
		if p == nil || p.Pool == nil {
			return
		}
		ctx := context.Background()
		exportName := handler
		req := &pluginv1.HandleRequest{Method: "CRON", Path: "/_cron/" + taskName}
		if _, err := callGuestFromPool(ctx, p, exportName, req,
			func() *pluginv1.HandleResponse { return &pluginv1.HandleResponse{} },
		); err != nil {
			logger.Errorf("plugin %q cron task %q error: %v", pluginName, taskName, err)
		}
	})
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.scheduledTasks = append(m.scheduledTasks, scheduledTask{
		pluginName: pluginName,
		taskName:   taskName,
		cronExpr:   cronExpr,
		handler:    handler,
		entryID:    entryID,
	})
	logger.Infof("plugin %q scheduled task %q: %s → %s", pluginName, taskName, cronExpr, handler)
	return nil
}

// onPluginEvent 当事件总线分发事件时调用
func (m *Manager) onPluginEvent(sub EventSubscription, event string, payload *structpb.Struct) {
	p := m.GetPlugin(sub.PluginName)
	if p == nil || p.Pool == nil {
		return
	}

	handlerName := strings.TrimSpace(sub.Handler)
	if handlerName == "" {
		handlerName = "on_" + sanitizeEventName(event)
	}

	ctx := context.Background()
	req := &pluginv1.EmitEventRequest{Event: event, Payload: payload}
	if _, err := callGuestFromPool(ctx, p, handlerName, req,
		func() *pluginv1.BoolResult { return &pluginv1.BoolResult{} },
	); err != nil {
		logger.Debugf("plugin %q event %q handler not found: %v", sub.PluginName, event, err)
	}
}

// SubscribeEvent 订阅宿主事件（供 host_subscribe_event 调用）
func (m *Manager) SubscribeEvent(pluginName, event, handler string) error {
	pluginName = strings.TrimSpace(pluginName)
	event = strings.TrimSpace(event)
	handler = strings.TrimSpace(handler)
	if pluginName == "" || event == "" {
		return fmt.Errorf("plugin name and event are required")
	}
	if handler == "" {
		handler = "on_" + sanitizeEventName(event)
	}
	m.eventBus.Subscribe(EventSubscription{
		PluginName: pluginName,
		Event:      event,
		Handler:    handler,
	})
	return nil
}

func (m *Manager) removeCronTasks(pluginName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var remainingTasks []scheduledTask
	for _, st := range m.scheduledTasks {
		if st.pluginName == pluginName {
			m.cronScheduler.Remove(st.entryID)
		} else {
			remainingTasks = append(remainingTasks, st)
		}
	}
	m.scheduledTasks = remainingTasks
}

func (m *Manager) removePluginRegistrations(pluginName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	filteredRoutes := m.pendingRoutes[:0]
	for _, route := range m.pendingRoutes {
		if route.PluginName == pluginName {
			delete(m.routes, routeKey(route.PluginName, route.Method, route.Path))
			continue
		}
		filteredRoutes = append(filteredRoutes, route)
	}
	m.pendingRoutes = filteredRoutes

	filteredMiddleware := m.pendingMiddleware[:0]
	for _, mw := range m.pendingMiddleware {
		if mw.PluginName == pluginName {
			delete(m.middleware, middlewareKey(mw.PluginName, mw.Name))
			continue
		}
		filteredMiddleware = append(filteredMiddleware, mw)
	}
	m.pendingMiddleware = filteredMiddleware
	m.eventBus.Unsubscribe(pluginName)
}

// ReloadPlugin 重载指定插件（停止→重新加载→启动）
func (m *Manager) ReloadPlugin(name string) error {
	name, err := normalizePluginName(name)
	if err != nil {
		return err
	}
	p := m.GetPlugin(name)
	if p == nil {
		return fmt.Errorf("plugin %q not found", name)
	}

	m.mu.Lock()
	delete(m.plugins, name)
	m.mu.Unlock()
	m.removeCronTasks(name)
	m.removePluginRegistrations(name)
	m.stopPlugin(context.Background(), p)

	// 重新扫描加载
	pluginPath, err := m.pluginPath(name)
	if err != nil {
		return err
	}
	manifest, err := ParseManifest(pluginPath)
	if err != nil {
		return fmt.Errorf("reload manifest: %w", err)
	}

	m.loadPlugin(context.Background(), pluginPath, manifest)
	return nil
}

// EnablePlugin 启用插件（加载并启动）
func (m *Manager) EnablePlugin(name string) error {
	name, err := normalizePluginName(name)
	if err != nil {
		return err
	}
	if existing := m.GetPlugin(name); existing != nil {
		return nil
	}

	pluginPath, err := m.pluginPath(name)
	if err != nil {
		return err
	}
	manifest, err := ParseManifest(pluginPath)
	if err != nil {
		return fmt.Errorf("enable plugin: %w", err)
	}

	m.loadPlugin(context.Background(), pluginPath, manifest)
	return nil
}

// DisablePlugin 禁用插件（停止并卸载）
func (m *Manager) DisablePlugin(name string) error {
	name, err := normalizePluginName(name)
	if err != nil {
		return err
	}
	p := m.GetPlugin(name)
	if p == nil {
		return fmt.Errorf("plugin %q not found", name)
	}

	m.mu.Lock()
	delete(m.plugins, name)
	m.mu.Unlock()
	m.removeCronTasks(name)
	m.removePluginRegistrations(name)
	m.stopPlugin(context.Background(), p)

	return nil
}

func routeKey(pluginName, method, path string) string {
	return pluginName + "\x00" + normalizeMethod(method) + "\x00" + normalizePath(path)
}

func middlewareKey(pluginName, name string) string {
	return pluginName + "\x00" + strings.TrimSpace(name)
}

func normalizeMethod(method string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "*" {
		return "ANY"
	}
	return method
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "*" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
		if path == "" {
			return "/"
		}
	}
	return path
}

// sanitizeEventName 将事件名转为合法的 WASM 导出函数名
func sanitizeEventName(event string) string {
	b := make([]byte, 0, len(event))
	for i := 0; i < len(event); i++ {
		c := event[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			b = append(b, c)
		} else {
			b = append(b, '_')
		}
	}
	return string(b)
}

// loadPlugins 加载所有插件
func (m *Manager) loadPlugins(ctx context.Context) {
	pluginDir := m.pluginDirectory()

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		logger.Errorf("create plugin directory %q: %v", pluginDir, err)
		return
	}

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		logger.Errorf("read plugin directory %q: %v", pluginDir, err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(pluginDir, entry.Name())

		manifest, err := ParseManifest(pluginPath)
		if err != nil {
			continue
		}

		if !m.isAllowed(manifest.Name) {
			continue
		}

		m.loadPlugin(ctx, pluginPath, manifest)
	}
}

func (m *Manager) isAllowed(name string) bool {
	for _, blocked := range m.config.BlockList {
		if blocked == name {
			return false
		}
	}
	if len(m.config.AllowList) > 0 {
		for _, allowed := range m.config.AllowList {
			if allowed == name {
				return true
			}
		}
		return false
	}
	return true
}

const (
	defaultPoolSize       = 1
	defaultPluginMemoryMB = 64
	defaultTimeoutSec     = 30
	defaultAsyncLimit     = 64
	wasmPageSize          = 64 * 1024
)

// RequestTimeout 返回插件请求/初始化默认超时。
func (m *Manager) RequestTimeout() time.Duration {
	sec := m.config.TimeoutSec
	if sec <= 0 {
		sec = defaultTimeoutSec
	}
	return time.Duration(sec) * time.Second
}

func (m *Manager) newPluginRuntime(ctx context.Context) (wazero.Runtime, error) {
	memoryMB := m.config.MaxMemoryMB
	if memoryMB <= 0 {
		memoryMB = defaultPluginMemoryMB
	}
	pages := uint32((memoryMB*1024*1024 + wasmPageSize - 1) / wasmPageSize)
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithMemoryLimitPages(pages).
		WithCloseOnContextDone(true),
	)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		_ = rt.Close(ctx)
		return nil, err
	}
	return rt, nil
}

func (m *Manager) loadPlugin(ctx context.Context, pluginPath string, manifest *PluginManifest) {
	instance := &PluginInstance{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Status:      StatusLoaded,
		Manifest:    manifest,
		AsyncNotify: make(chan uint64, 1024),
		AsyncSem:    make(chan struct{}, defaultAsyncLimit),
	}

	rt, err := m.newPluginRuntime(ctx)
	if err != nil {
		instance.SetError("runtime: " + err.Error())
		m.mu.Lock()
		m.plugins[manifest.Name] = instance
		m.mu.Unlock()
		return
	}
	instance.Runtime = rt

	wasmFile := manifest.Main
	if wasmFile == "" {
		wasmFile = "plugin.wasm"
	}
	compiled, err := CompilePlugin(ctx, rt, pluginPath, wasmFile)
	if err != nil {
		instance.SetError("compile: " + err.Error())
		_ = rt.Close(ctx)
		instance.Runtime = nil
		m.mu.Lock()
		m.plugins[manifest.Name] = instance
		m.mu.Unlock()
		return
	}

	hostFuncs := m.buildHostFunctions(manifest.Name, manifest)

	poolSize := defaultPoolSize
	pool, err := InstantiateWithHostFuncs(ctx, rt, compiled, manifest.Name, hostFuncs, poolSize)
	if err != nil {
		instance.SetError("instantiate: " + err.Error())
		_ = rt.Close(ctx)
		instance.Runtime = nil
		m.mu.Lock()
		m.plugins[manifest.Name] = instance
		m.mu.Unlock()
		return
	}

	instance.Pool = pool
	instance.SetStatus(StatusInitialized)
	m.mu.Lock()
	m.plugins[manifest.Name] = instance
	m.mu.Unlock()

	initReq := m.buildInitRequest(manifest)
	initCtx, cancel := context.WithTimeout(ctx, m.RequestTimeout())
	defer cancel()
	for i := 0; i < poolSize; i++ {
		mod, err := instance.Pool.Get(initCtx)
		if err != nil {
			m.failLoadedPlugin(ctx, instance, "init acquire instance: "+err.Error())
			return
		}
		res, err := CallGuestFn(initCtx, mod, "init", initReq,
			func() *pluginv1.BoolResult { return &pluginv1.BoolResult{} },
		)
		if err != nil {
			instance.Pool.Put(mod)
			m.failLoadedPlugin(ctx, instance, "init: "+err.Error())
			return
		}
		if res != nil && !res.Success {
			instance.Pool.Put(mod)
			m.failLoadedPlugin(ctx, instance, "init: "+res.Error)
			return
		}

		if err := CallGuestStart(initCtx, mod); err != nil {
			instance.Pool.Put(mod)
			m.failLoadedPlugin(ctx, instance, "start: "+err.Error())
			return
		}
		instance.Pool.Put(mod)
	}

	instance.SetStatus(StatusRunning)
}

func (m *Manager) failLoadedPlugin(ctx context.Context, instance *PluginInstance, msg string) {
	instance.SetError(msg)
	m.removeCronTasks(instance.Name)
	m.removePluginRegistrations(instance.Name)
	if instance.Pool != nil {
		instance.Pool.Close(ctx)
		instance.Pool = nil
	}
	if instance.Runtime != nil {
		_ = instance.Runtime.Close(ctx)
		instance.Runtime = nil
	}
}

// buildHostFunctions 为插件构建宿主函数集合
// 根据 manifest.Permissions 按需注册
func (m *Manager) buildHostFunctions(pluginName string, manifest *PluginManifest) map[string]api.GoModuleFunc {
	funcs := make(map[string]api.GoModuleFunc)

	// host_log — 总是可用
	funcs["host_log"] = wrapHostLog(pluginName)

	// 按权限注册其他宿主函数
	for _, perm := range manifest.Permissions {
		switch Permission(perm) {
		case PermHTTPRoutes:
			funcs["host_register_route"] = wrapHostRegisterRoute(pluginName, m)
		case PermMiddleware:
			funcs["host_register_middleware"] = wrapHostRegisterMiddleware(pluginName, m)
		case PermRedis:
			funcs["host_redis_get"] = wrapHostRedisGet(pluginName, m)
			funcs["host_redis_set"] = wrapHostRedisSet(pluginName, m)
		case PermConfigRead:
			funcs["host_config_get"] = wrapHostConfigGet(pluginName, m)
		case PermEvents:
			funcs["host_emit_event"] = wrapHostEmitEvent(pluginName, m)
			funcs["host_subscribe_event"] = wrapHostSubscribeEvent(pluginName, m)
		case PermHTTPClient:
			funcs["host_http_request"] = wrapHostHTTPRequest(pluginName, m)
		case PermScheduler:
			funcs["host_schedule_task"] = wrapHostScheduleTask(pluginName, m)
		case PermDatabaseRead, PermDatabaseWrite:
			funcs["host_db_query"] = wrapHostDBQuery(pluginName, m)
		case PermQueue:
			funcs["host_enqueue_task"] = wrapHostEnqueueTask(pluginName, m)
		}
	}

	// 异步运行时函数（总是可用）
	funcs["host_async_submit"] = wrapHostAsyncSubmit(pluginName, m)
	funcs["host_async_resolve"] = wrapHostAsyncResolve(pluginName, m)
	funcs["host_async_wait_any"] = wrapHostAsyncWaitAny(pluginName, m)

	return funcs
}

// ============================================================================
// Async Task Management — goroutine 池
// ============================================================================

// submitAsyncTask 提交异步任务到 goroutine 池
func (m *Manager) submitAsyncTask(pluginName, opType string, params []byte) uint64 {
	p := m.GetPlugin(pluginName)
	if p == nil || p.AsyncNotify == nil || p.AsyncSem == nil {
		return 0
	}
	select {
	case p.AsyncSem <- struct{}{}:
	default:
		return 0
	}

	m.asyncTaskMu.Lock()
	m.asyncTaskSeq++
	id := m.asyncTaskSeq
	task := &asyncTask{
		id:         id,
		pluginName: pluginName,
		opType:     opType,
		params:     params,
		done:       make(chan struct{}),
		notifyCh:   p.AsyncNotify,
	}
	if p.AsyncSem != nil {
		task.release = func() { <-p.AsyncSem }
	}
	m.asyncTasks[id] = task
	m.asyncTaskMu.Unlock()

	// 在后台 goroutine 执行实际操作
	go m.executeAsyncTask(task)

	return id
}

// executeAsyncTask 在 goroutine 中执行异步任务。
// task.params 现在是 protobuf 编码的请求消息（与同步宿主函数一致）。
func (m *Manager) executeAsyncTask(task *asyncTask) {
	defer func() {
		if task.release != nil {
			task.release()
		}
		close(task.done)
	}()

	ctx := context.Background()

	switch task.opType {
	case "http_get", "http_post":
		if !m.HasPermission(task.pluginName, PermHTTPClient) {
			task.err = "permission denied: missing http_client"
			break
		}
		method := "GET"
		if task.opType == "http_post" {
			method = "POST"
		}
		req := &pluginv1.HttpRequestRequest{Method: method}
		if err := proto.Unmarshal(task.params, req); err != nil {
			task.err = "decode " + task.opType + " params: " + err.Error()
		} else {
			if !isURLAllowed(req.Url) {
				resp := &pluginv1.HttpRequestResponse{Status: 403, Body: []byte("URL blocked")}
				task.result, _ = proto.Marshal(resp)
			} else {
				resp, err := doHTTP(ctx, req)
				if err != nil {
					task.err = err.Error()
				} else {
					task.result, _ = proto.Marshal(resp)
				}
			}
		}

	case "redis_get":
		if !m.HasPermission(task.pluginName, PermRedis) {
			task.err = "permission denied: missing redis"
			break
		}
		req := &pluginv1.RedisGetRequest{}
		if err := proto.Unmarshal(task.params, req); err != nil {
			task.err = "decode redis_get params: " + err.Error()
		} else {
			key := fmt.Sprintf("plugin:%s:%s", task.pluginName, req.Key)
			val, err := m.hostEnv.Redis.Get(key)
			if err != nil {
				task.err = err.Error()
			} else {
				resp := &pluginv1.RedisGetResponse{Value: val, Exists: true}
				task.result, _ = proto.Marshal(resp)
			}
		}

	case "redis_set":
		if !m.HasPermission(task.pluginName, PermRedis) {
			task.err = "permission denied: missing redis"
			break
		}
		req := &pluginv1.RedisSetRequest{}
		if err := proto.Unmarshal(task.params, req); err != nil {
			task.err = "decode redis_set params: " + err.Error()
		} else {
			key := fmt.Sprintf("plugin:%s:%s", task.pluginName, req.Key)
			err := m.hostEnv.Redis.Set(key, req.Value, req.TtlSeconds)
			if err != nil {
				task.err = err.Error()
			} else {
				resp := &pluginv1.BoolResult{Success: true}
				task.result, _ = proto.Marshal(resp)
			}
		}

	case "db_query":
		req := &pluginv1.DbQueryRequest{}
		if err := proto.Unmarshal(task.params, req); err != nil {
			task.err = "decode db_query params: " + err.Error()
		} else if m.hostEnv.Store != nil {
			isWrite := req.Operation == "create" || req.Operation == "update" || req.Operation == "delete"
			if isWrite && !m.HasPermission(task.pluginName, PermDatabaseWrite) {
				task.err = "permission denied: missing database_write"
			} else if !isWrite && !m.HasPermission(task.pluginName, PermDatabaseRead) && !m.HasPermission(task.pluginName, PermDatabaseWrite) {
				task.err = "permission denied: missing database_read"
			} else {
				conditions := structToMap(req.Conditions)
				fields := make([]string, 0)
				if req.Fields != nil {
					fields = req.Fields.Paths
				}
				rows, total, err := m.hostEnv.Store.Query(req.Model, req.Operation, conditions, fields, req.Limit, req.Offset)
				if err != nil {
					task.err = err.Error()
				} else {
					pbRows := make([]*structpb.Struct, 0, len(rows))
					for _, row := range rows {
						if s, err := structpb.NewStruct(toStringMap(row)); err == nil {
							pbRows = append(pbRows, s)
						}
					}
					resp := &pluginv1.DbQueryResponse{Rows: pbRows, Total: total}
					task.result, _ = proto.Marshal(resp)
				}
			}
		}
	}

	if task.notifyCh != nil {
		select {
		case task.notifyCh <- task.id:
		default:
			// 队列满，丢弃通知并清理内存
			m.asyncTaskMu.Lock()
			delete(m.asyncTasks, task.id)
			m.asyncTaskMu.Unlock()
		}
	} else {
		m.asyncTaskMu.Lock()
		delete(m.asyncTasks, task.id)
		m.asyncTaskMu.Unlock()
	}
}

// resolveAsyncTask 阻塞直到指定异步任务完成，返回结果
func (m *Manager) resolveAsyncTask(pluginName string, id uint64) ([]byte, string, bool) {
	m.asyncTaskMu.Lock()
	task, ok := m.asyncTasks[id]
	m.asyncTaskMu.Unlock()
	if !ok || task.pluginName != pluginName {
		return nil, "task not found", true
	}

	<-task.done

	m.asyncTaskMu.Lock()
	delete(m.asyncTasks, id)
	m.asyncTaskMu.Unlock()

	return task.result, task.err, true
}

// waitAnyAsyncTask 阻塞直到任意一个未完成的异步任务完成
func (m *Manager) waitAnyAsyncTask(pluginName string) uint64 {
	p := m.GetPlugin(pluginName)
	if p == nil || p.AsyncNotify == nil {
		return 0
	}
	return <-p.AsyncNotify
}

func splitParams(data []byte, n int) []string {
	s := string(data)
	parts := make([]string, 0, n)
	for i := 0; i < n-1; i++ {
		idx := indexByte(s, '|')
		if idx < 0 {
			parts = append(parts, s)
			return parts
		}
		parts = append(parts, s[:idx])
		s = s[idx+1:]
	}
	if s != "" {
		parts = append(parts, s)
	}
	return parts
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func parseInt64(s string) (int64, error) {
	var n int64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return n, fmt.Errorf("invalid number: %s", s)
		}
		n = n*10 + int64(s[i]-'0')
	}
	return n, nil
}

func jsonMarshal(v interface{}) ([]byte, error) {
	// minimal JSON marshaling for map types
	switch val := v.(type) {
	case map[string]interface{}:
		return []byte(fmt.Sprintf("%v", val)), nil
	}
	return nil, fmt.Errorf("unsupported type")
}

func (m *Manager) doHTTPRequest(ctx context.Context, method string, params []byte) ([]byte, error) {
	// params format: url|body
	parts := splitParams(params, 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("url required")
	}
	url := parts[0]
	var body []byte
	if len(parts) >= 2 {
		body = []byte(parts[1])
	}
	resp, err := doHTTP(ctx, &pluginv1.HttpRequestRequest{
		Method: method,
		Url:    url,
		Body:   body,
	})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(resp)
}
