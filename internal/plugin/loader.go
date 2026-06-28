package plugin

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// CompilePlugin 从文件编译 WASM 模块
func CompilePlugin(ctx context.Context, rt wazero.Runtime, pluginPath, wasmFile string) (wazero.CompiledModule, error) {
	wasmPath, err := ResolvePluginFile(pluginPath, wasmFile)
	if err != nil {
		return nil, err
	}
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("read wasm file %s: %w", wasmPath, err)
	}

	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("compile wasm module: %w", err)
	}

	return compiled, nil
}

// ResolvePluginFile 解析插件目录内的文件路径，拒绝绝对路径和路径穿越。
func ResolvePluginFile(pluginPath, file string) (string, error) {
	file = strings.TrimSpace(file)
	if file == "" {
		return "", fmt.Errorf("plugin file is required")
	}
	if filepath.IsAbs(file) {
		return "", fmt.Errorf("plugin file must be relative: %s", file)
	}
	cleaned := filepath.Clean(file)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("plugin file escapes plugin directory: %s", file)
	}

	baseAbs, err := filepath.Abs(pluginPath)
	if err != nil {
		return "", fmt.Errorf("resolve plugin directory: %w", err)
	}
	targetAbs, err := filepath.Abs(filepath.Join(pluginPath, cleaned))
	if err != nil {
		return "", fmt.Errorf("resolve plugin file: %w", err)
	}
	if targetAbs != baseAbs && !strings.HasPrefix(targetAbs, baseAbs+string(os.PathSeparator)) {
		return "", fmt.Errorf("plugin file escapes plugin directory: %s", file)
	}
	return targetAbs, nil
}

// NewModuleConfig 创建模块配置
func NewModuleConfig(pluginName string) wazero.ModuleConfig {
	dataDir := filepath.Join("data", "plugins", pluginName)
	_ = os.MkdirAll(dataDir, 0755)

	fsConfig := wazero.NewFSConfig().WithDirMount(dataDir, "/data")

	return wazero.NewModuleConfig().
		WithName(pluginName).
		WithSysWalltime().
		WithSysNanotime().
		WithRandSource(rand.Reader).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithFSConfig(fsConfig)
}

// InstantiateWithHostFuncs 注册宿主函数并实例化 WASM 模块
//
// hostFuncs: 宿主函数名 → api.GoModuleFunc 的映射
func InstantiateWithHostFuncs(
	ctx context.Context,
	rt wazero.Runtime,
	compiled wazero.CompiledModule,
	pluginName string,
	hostFuncs map[string]api.GoModuleFunc,
	poolSize int,
) (*InstancePool, error) {

	// 创建宿主模块，注册所有宿主函数
	hostBuilder := rt.NewHostModuleBuilder("env")
	for name, fn := range hostFuncs {
		hostBuilder = hostBuilder.NewFunctionBuilder().
			WithGoModuleFunction(fn, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI64}).
			Export(name)
	}

	if _, err := hostBuilder.Instantiate(ctx); err != nil {
		return nil, fmt.Errorf("instantiate host module: %w", err)
	}

	// 实例化插件模块
	pool := NewInstancePool(poolSize)
	for i := 0; i < poolSize; i++ {
		cfg := NewModuleConfig(fmt.Sprintf("%s_%d", pluginName, i))
		mod, err := rt.InstantiateModule(ctx, compiled, cfg)
		if err != nil {
			pool.Close(ctx)
			return nil, fmt.Errorf("instantiate plugin module %d: %w", i, err)
		}
		pool.Put(mod)
	}

	return pool, nil
}

// CloseModule 安全关闭模块
func CloseModule(ctx context.Context, mod api.Module) error {
	if mod != nil {
		return mod.Close(ctx)
	}
	return nil
}
