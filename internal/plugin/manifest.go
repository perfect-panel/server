package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PluginManifest 是插件描述文件 plugin.yaml 的结构
type PluginManifest struct {
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Description string                 `yaml:"description"`
	Author      string                 `yaml:"author"`
	Main        string                 `yaml:"main"` // .wasm 文件路径，相对于插件目录
	Permissions []string               `yaml:"permissions"`
	Config      map[string]interface{} `yaml:"config"`
}

// ParseManifest 从插件目录解析 plugin.yaml 并校验
func ParseManifest(pluginDir string) (*PluginManifest, error) {
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}

	var m PluginManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", manifestPath, err)
	}

	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("validate manifest %s: %w", manifestPath, err)
	}

	return &m, nil
}

// Validate 校验清单的必填字段和权限合法性
func (m *PluginManifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Version == "" {
		return fmt.Errorf("version is required")
	}
	if m.Main == "" {
		return fmt.Errorf("main (wasm file) is required")
	}

	// 校验权限
	for _, p := range m.Permissions {
		if !AllowedPermissions[Permission(p)] {
			return fmt.Errorf("unknown permission: %s", p)
		}
	}

	return nil
}

// HasPermission 检查插件是否声明了某个权限
func (m *PluginManifest) HasPermission(perm Permission) bool {
	for _, p := range m.Permissions {
		if Permission(p) == perm {
			return true
		}
	}
	return false
}
