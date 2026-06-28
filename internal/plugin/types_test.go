package plugin

import (
	"sync"
	"testing"
)

func TestPluginInstanceSetGetStatus(t *testing.T) {
	p := &PluginInstance{Status: StatusUnloaded}
	p.SetStatus(StatusLoaded)
	if s := p.GetStatus(); s != StatusLoaded {
		t.Errorf("status = %q, want %q", s, StatusLoaded)
	}
}

func TestPluginInstanceConcurrentStatus(t *testing.T) {
	p := &PluginInstance{Status: StatusUnloaded}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.SetStatus(StatusRunning)
			_ = p.GetStatus()
		}()
	}
	wg.Wait()
}

func TestPluginInstanceAddGetRoutes(t *testing.T) {
	p := &PluginInstance{Name: "test"}
	p.AddRoute(RouteRegistration{PluginName: "test", Method: "GET", Path: "/hello", Handler: "h"})
	p.AddRoute(RouteRegistration{PluginName: "test", Method: "POST", Path: "/submit", Handler: "s"})

	routes := p.GetRoutes()
	if len(routes) != 2 {
		t.Fatalf("routes len = %d, want 2", len(routes))
	}
	if routes[0].Path != "/hello" {
		t.Errorf("route[0].path = %q, want %q", routes[0].Path, "/hello")
	}
	if routes[1].Method != "POST" {
		t.Errorf("route[1].method = %q, want %q", routes[1].Method, "POST")
	}
}

func TestPluginInstanceSetError(t *testing.T) {
	p := &PluginInstance{Status: StatusLoaded}
	p.SetError("something went wrong")
	if p.Error != "something went wrong" {
		t.Errorf("error = %q", p.Error)
	}
	if p.Status != StatusError {
		t.Errorf("status = %q, want %q", p.Status, StatusError)
	}
}

func TestPluginInstanceToInfo(t *testing.T) {
	p := &PluginInstance{
		Name:        "test-plugin",
		Version:     "1.0",
		Description: "A test",
		Author:      "dev",
		Status:      StatusRunning,
		Manifest: &PluginManifest{
			Permissions: []string{"http_routes", "logging"},
		},
	}
	p.AddRoute(RouteRegistration{Method: "GET", Path: "/hello"})

	info := p.ToInfo()
	if info.Name != "test-plugin" {
		t.Errorf("name = %q", info.Name)
	}
	if info.Status != StatusRunning {
		t.Errorf("status = %q", info.Status)
	}
	if len(info.Permissions) != 2 {
		t.Errorf("permissions = %d", len(info.Permissions))
	}
	if len(info.Routes) != 1 {
		t.Errorf("routes = %d", len(info.Routes))
	}
	if info.Routes[0] != "GET /hello" {
		t.Errorf("routes[0] = %q", info.Routes[0])
	}
}

func TestPermissionConstants(t *testing.T) {
	// Verify all constants are unique
	seen := make(map[Permission]bool)
	for _, p := range []Permission{
		PermHTTPRoutes, PermMiddleware, PermDatabaseRead, PermDatabaseWrite,
		PermRedis, PermLogging, PermConfigRead, PermEvents, PermHTTPClient, PermScheduler, PermQueue,
	} {
		if seen[p] {
			t.Errorf("duplicate permission: %s", p)
		}
		seen[p] = true
	}

	// Verify all permissions are in AllowedPermissions
	for p := range seen {
		if !AllowedPermissions[p] {
			t.Errorf("permission %s not in AllowedPermissions", p)
		}
	}
}
