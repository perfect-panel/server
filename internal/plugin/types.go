package plugin

import (
	"context"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Permission 定义插件可以请求的能力
type Permission string

const (
	PermHTTPRoutes    Permission = "http_routes"
	PermMiddleware    Permission = "middleware"
	PermDatabaseRead  Permission = "database_read"
	PermDatabaseWrite Permission = "database_write"
	PermRedis         Permission = "redis"
	PermLogging       Permission = "logging"
	PermConfigRead    Permission = "config_read"
	PermEvents        Permission = "events"
	PermHTTPClient    Permission = "http_client"
	PermScheduler     Permission = "scheduler"
	PermQueue         Permission = "queue"
)

// AllowedPermissions 定义所有合法权限的集合
var AllowedPermissions = map[Permission]bool{
	PermHTTPRoutes:    true,
	PermMiddleware:    true,
	PermDatabaseRead:  true,
	PermDatabaseWrite: true,
	PermRedis:         true,
	PermLogging:       true,
	PermConfigRead:    true,
	PermEvents:        true,
	PermHTTPClient:    true,
	PermScheduler:     true,
	PermQueue:         true,
}

// PluginStatus 表示插件的当前状态
type PluginStatus string

const (
	StatusUnloaded    PluginStatus = "unloaded"
	StatusLoaded      PluginStatus = "loaded"
	StatusInitialized PluginStatus = "initialized"
	StatusRunning     PluginStatus = "running"
	StatusStopped     PluginStatus = "stopped"
	StatusError       PluginStatus = "error"
)

// RouteRegistration 插件注册的路由
type RouteRegistration struct {
	PluginName string   // 插件名
	Method     string   // GET, POST, PUT, DELETE
	Path       string   // 插件自定义路径，如 "/stats"
	Handler    string   // WASM 导出函数名
	Middleware []string // "auth", "device"
}

// MiddlewareRegistration 插件注册的中间件
type MiddlewareRegistration struct {
	PluginName string // 插件名
	Name       string // 中间件名（用于路由引用）
	Handler    string // WASM 导出函数名，如 "mw_rate_limit"
}

// PluginInfo 对外暴露的插件信息（用于管理 API）
type PluginInfo struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	Author      string       `json:"author"`
	Status      PluginStatus `json:"status"`
	Permissions []string     `json:"permissions"`
	Routes      []string     `json:"routes"` // 已注册的路由路径列表
	Error       string       `json:"error,omitempty"`
}

// PluginHealth 描述插件运行时健康状态（用于管理 API）。
type PluginHealth struct {
	Name            string       `json:"name"`
	Status          PluginStatus `json:"status"`
	Ready           bool         `json:"ready"`
	PoolSize        int          `json:"pool_size"`
	AsyncInFlight   int          `json:"async_in_flight"`
	AsyncLimit      int          `json:"async_limit"`
	RegisteredRoute int          `json:"registered_route"`
	Error           string       `json:"error,omitempty"`
}

// PluginValidationCheck 表示插件安装校验中的单项结果。
type PluginValidationCheck struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

// PluginValidation 表示插件安装包/目录的校验结果。
type PluginValidation struct {
	Name     string                  `json:"name"`
	Valid    bool                    `json:"valid"`
	Manifest *PluginManifest         `json:"manifest,omitempty"`
	Checks   []PluginValidationCheck `json:"checks"`
	Error    string                  `json:"error,omitempty"`
}

// HostFuncDef 定义宿主函数的元数据
type HostFuncDef struct {
	Name       string      // WASM 导入函数名
	Fn         interface{} // api.GoModuleFunc
	Permission Permission  // 所需权限
}

// EventSubscription 事件订阅
type EventSubscription struct {
	PluginName string // 订阅者插件名
	Event      string // 事件名
	Handler    string // WASM 导出函数名，如 "on_order_paid"
}

// ManagerStatus 管理器状态
type ManagerStatus int

const (
	ManagerUninitialized ManagerStatus = iota
	ManagerLoading
	ManagerRunning
	ManagerStopping
	ManagerStopped
)

// PluginInstance 代表一个已加载的插件实例
type PluginInstance struct {
	mu sync.RWMutex

	Name        string
	Version     string
	Description string
	Author      string
	Status      PluginStatus
	Manifest    *PluginManifest
	Routes      []RouteRegistration
	Error       string

	// Pool 管理该插件的 WASM 模块实例
	Pool *InstancePool

	// Runtime 是该插件独立的 wazero runtime，避免 env host module 跨插件冲突
	Runtime wazero.Runtime

	// AsyncNotify 异步任务完成通知
	AsyncNotify chan uint64

	// AsyncSem 限制该插件未完成的异步任务数
	AsyncSem chan struct{}
}

// InstancePool 管理 WASM 模块实例池
type InstancePool struct {
	pool   chan api.Module
	size   int
	closed bool
	mu     sync.Mutex
}

// NewInstancePool 创建一个新的实例池
func NewInstancePool(size int) *InstancePool {
	return &InstancePool{
		pool: make(chan api.Module, size),
		size: size,
	}
}

// Size 返回池的大小
func (p *InstancePool) Size() int {
	return p.size
}

// Get 获取一个空闲的 WASM 实例
func (p *InstancePool) Get(ctx context.Context) (api.Module, error) {
	select {
	case mod, ok := <-p.pool:
		if !ok {
			return nil, context.Canceled
		}
		return mod, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Put 归还 WASM 实例
func (p *InstancePool) Put(mod api.Module) {
	if mod == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		mod.Close(context.Background())
		return
	}
	select {
	case p.pool <- mod:
	default:
		mod.Close(context.Background())
	}
}

// Close 关闭池中所有实例
func (p *InstancePool) Close(ctx context.Context) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	close(p.pool)
	p.mu.Unlock()

	for mod := range p.pool {
		if mod != nil {
			mod.Close(ctx)
		}
	}
}

// SetStatus 安全设置插件状态
func (p *PluginInstance) SetStatus(s PluginStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Status = s
}

// GetStatus 安全获取插件状态
func (p *PluginInstance) GetStatus() PluginStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Status
}

// SetError 安全设置错误信息
func (p *PluginInstance) SetError(err string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Error = err
	p.Status = StatusError
}

// AddRoute 安全添加路由
func (p *PluginInstance) AddRoute(route RouteRegistration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Routes = append(p.Routes, route)
}

// GetRoutes 安全获取路由列表
func (p *PluginInstance) GetRoutes() []RouteRegistration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]RouteRegistration, len(p.Routes))
	copy(result, p.Routes)
	return result
}

// ToInfo 转换为对外暴露的 PluginInfo
func (p *PluginInstance) ToInfo() PluginInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	routes := make([]string, len(p.Routes))
	for i, r := range p.Routes {
		routes[i] = r.Method + " " + r.Path
	}

	perms := make([]string, 0)
	if p.Manifest != nil {
		for _, perm := range p.Manifest.Permissions {
			perms = append(perms, string(perm))
		}
	}

	return PluginInfo{
		Name:        p.Name,
		Version:     p.Version,
		Description: p.Description,
		Author:      p.Author,
		Status:      p.Status,
		Permissions: perms,
		Routes:      routes,
		Error:       p.Error,
	}
}
