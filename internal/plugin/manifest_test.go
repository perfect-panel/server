package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseManifestValid(t *testing.T) {
	dir := t.TempDir()
	yaml := `name: test-plugin
version: 1.0.0
description: A test plugin
author: tester
main: plugin.wasm
permissions:
  - http_routes
  - logging
config:
  key: value
`
	os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte(yaml), 0644)

	m, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "test-plugin" {
		t.Errorf("name = %q, want %q", m.Name, "test-plugin")
	}
	if m.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", m.Version, "1.0.0")
	}
	if m.Main != "plugin.wasm" {
		t.Errorf("main = %q, want %q", m.Main, "plugin.wasm")
	}
	if len(m.Permissions) != 2 {
		t.Errorf("permissions len = %d, want 2", len(m.Permissions))
	}
}

func TestParseManifestMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ParseManifest(dir)
	if err == nil {
		t.Fatal("expected error for missing plugin.yaml")
	}
}

func TestParseManifestInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte("{{{bad yaml"), 0644)
	_, err := ParseManifest(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestValidateMissingName(t *testing.T) {
	m := &PluginManifest{Version: "1.0", Main: "plugin.wasm"}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidateMissingVersion(t *testing.T) {
	m := &PluginManifest{Name: "test", Main: "plugin.wasm"}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestValidateMissingMain(t *testing.T) {
	m := &PluginManifest{Name: "test", Version: "1.0"}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for missing main")
	}
}

func TestValidateUnknownPermission(t *testing.T) {
	m := &PluginManifest{
		Name:        "test",
		Version:     "1.0",
		Main:        "plugin.wasm",
		Permissions: []string{"unknown_permission"},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for unknown permission")
	}
}

func TestValidateAllKnownPermissions(t *testing.T) {
	perms := []string{
		"http_routes", "middleware", "database_read", "database_write",
		"redis", "logging", "config_read", "events", "http_client", "scheduler",
		"queue",
	}
	m := &PluginManifest{
		Name: "test", Version: "1.0", Main: "plugin.wasm",
		Permissions: perms,
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("unexpected error for known permissions: %v", err)
	}
}

func TestHasPermission(t *testing.T) {
	m := &PluginManifest{Permissions: []string{"http_routes", "logging"}}
	if !m.HasPermission(PermHTTPRoutes) {
		t.Error("should have http_routes")
	}
	if !m.HasPermission(PermLogging) {
		t.Error("should have logging")
	}
	if m.HasPermission(PermRedis) {
		t.Error("should not have redis")
	}
}
