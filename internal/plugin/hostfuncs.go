package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	pluginv1 "github.com/perfect-panel/server/api/plugin/v1"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/tetratelabs/wazero/api"
	"google.golang.org/protobuf/types/known/structpb"
)

// ============================================================================
// Host Function Wrappers — 每个宿主函数都遵循相同的 MakeHostFunc 模式
// ============================================================================

// host_log — 写日志
func wrapHostLog(pluginName string) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.LogRequest) (*pluginv1.BoolResult, error) {
			msg := fmt.Sprintf("[plugin:%s] %s", name, req.Message)
			switch req.Level {
			case "info":
				logger.Info(msg)
			case "warn":
				logger.Infof("WARN %s", msg) // logger 无 Warn 级别，用 Info 替代
			case "error":
				logger.Error(msg)
			case "debug":
				logger.Debug(msg)
			default:
				logger.Info(msg)
			}
			return &pluginv1.BoolResult{Success: true}, nil
		},
		func() *pluginv1.LogRequest { return &pluginv1.LogRequest{} },
	)
}

// host_config_get — 读配置
func wrapHostConfigGet(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.ConfigGetRequest) (*pluginv1.ConfigGetResponse, error) {
			key := strings.TrimSpace(req.Key)
			if key == "" {
				return &pluginv1.ConfigGetResponse{Exists: false}, nil
			}
			if !isSafeConfigKey(key) {
				return &pluginv1.ConfigGetResponse{Exists: false}, nil
			}
			val := m.GetConfigValue(key)
			return &pluginv1.ConfigGetResponse{Value: val, Exists: val != ""}, nil
		},
		func() *pluginv1.ConfigGetRequest { return &pluginv1.ConfigGetRequest{} },
	)
}

// host_register_route — 注册 HTTP 路由
func wrapHostRegisterRoute(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.RegisterRouteRequest) (*pluginv1.BoolResult, error) {
			if err := m.RegisterRoute(RouteRegistration{
				PluginName: name,
				Method:     req.Method,
				Path:       req.Path,
				Handler:    req.Handler,
				Middleware: req.Middleware,
			}); err != nil {
				return &pluginv1.BoolResult{Success: false, Error: err.Error()}, nil
			}
			return &pluginv1.BoolResult{Success: true}, nil
		},
		func() *pluginv1.RegisterRouteRequest { return &pluginv1.RegisterRouteRequest{} },
	)
}

// host_redis_get — 读取 Redis
func wrapHostRedisGet(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.RedisGetRequest) (*pluginv1.RedisGetResponse, error) {
			key := fmt.Sprintf("plugin:%s:%s", name, req.Key)
			if m.GetRedis() == nil {
				return &pluginv1.RedisGetResponse{Exists: false}, nil
			}
			val, err := m.GetRedis().Get(key)
			if err != nil {
				return &pluginv1.RedisGetResponse{Exists: false}, nil
			}
			return &pluginv1.RedisGetResponse{Value: val, Exists: true}, nil
		},
		func() *pluginv1.RedisGetRequest { return &pluginv1.RedisGetRequest{} },
	)
}

// host_redis_set — 写入 Redis
func wrapHostRedisSet(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.RedisSetRequest) (*pluginv1.BoolResult, error) {
			key := fmt.Sprintf("plugin:%s:%s", name, req.Key)
			if m.GetRedis() == nil {
				return &pluginv1.BoolResult{Success: false, Error: "redis not available"}, nil
			}
			err := m.GetRedis().Set(key, req.Value, req.TtlSeconds)
			if err != nil {
				return &pluginv1.BoolResult{Success: false, Error: err.Error()}, nil
			}
			return &pluginv1.BoolResult{Success: true}, nil
		},
		func() *pluginv1.RedisSetRequest { return &pluginv1.RedisSetRequest{} },
	)
}

// host_emit_event — 发布事件
func wrapHostEmitEvent(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.EmitEventRequest) (*pluginv1.BoolResult, error) {
			m.EventBus().Publish(req.Event, req.Payload)
			return &pluginv1.BoolResult{Success: true}, nil
		},
		func() *pluginv1.EmitEventRequest { return &pluginv1.EmitEventRequest{} },
	)
}

// host_subscribe_event — 订阅事件
func wrapHostSubscribeEvent(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.SubscribeEventRequest) (*pluginv1.BoolResult, error) {
			if err := m.SubscribeEvent(name, req.Event, req.Handler); err != nil {
				return &pluginv1.BoolResult{Success: false, Error: err.Error()}, nil
			}
			return &pluginv1.BoolResult{Success: true}, nil
		},
		func() *pluginv1.SubscribeEventRequest { return &pluginv1.SubscribeEventRequest{} },
	)
}

// host_db_query — 数据库查询
func wrapHostDBQuery(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.DbQueryRequest) (*pluginv1.DbQueryResponse, error) {
			isWrite := req.Operation == "create" || req.Operation == "update" || req.Operation == "delete"
			if isWrite && !m.HasPermission(name, PermDatabaseWrite) {
				return &pluginv1.DbQueryResponse{Error: "permission denied: missing database_write"}, nil
			} else if !isWrite && !m.HasPermission(name, PermDatabaseRead) && !m.HasPermission(name, PermDatabaseWrite) {
				return &pluginv1.DbQueryResponse{Error: "permission denied: missing database_read"}, nil
			}

			conditions := structToMap(req.Conditions)
			fields := make([]string, 0)
			if req.Fields != nil {
				fields = req.Fields.Paths
			}

			store := m.GetStore()
			if store == nil {
				return &pluginv1.DbQueryResponse{Error: "database not available"}, nil
			}
			rows, total, err := store.Query(req.Model, req.Operation, conditions, fields, req.Limit, req.Offset)
			if err != nil {
				return &pluginv1.DbQueryResponse{Error: err.Error()}, nil
			}

			pbRows := make([]*structpb.Struct, 0, len(rows))
			for _, row := range rows {
				if s, err := structpb.NewStruct(toStringMap(row)); err == nil {
					pbRows = append(pbRows, s)
				}
			}

			return &pluginv1.DbQueryResponse{Rows: pbRows, Total: total}, nil
		},
		func() *pluginv1.DbQueryRequest { return &pluginv1.DbQueryRequest{} },
	)
}

// host_http_request — HTTP 客户端（防 SSRF）
func wrapHostHTTPRequest(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.HttpRequestRequest) (*pluginv1.HttpRequestResponse, error) {
			if !isURLAllowed(req.Url) {
				return &pluginv1.HttpRequestResponse{Status: 403, Body: []byte("URL blocked")}, nil
			}
			return doHTTP(ctx, req)
		},
		func() *pluginv1.HttpRequestRequest { return &pluginv1.HttpRequestRequest{} },
	)
}

// host_schedule_task — 定时任务
func wrapHostScheduleTask(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.ScheduleTaskRequest) (*pluginv1.BoolResult, error) {
			if err := m.ScheduleTask(name, req.Name, req.Cron, req.Handler); err != nil {
				return &pluginv1.BoolResult{Success: false, Error: err.Error()}, nil
			}
			return &pluginv1.BoolResult{Success: true}, nil
		},
		func() *pluginv1.ScheduleTaskRequest { return &pluginv1.ScheduleTaskRequest{} },
	)
}

// host_register_middleware — 注册中间件
func wrapHostRegisterMiddleware(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.RegisterMiddlewareRequest) (*pluginv1.BoolResult, error) {
			if err := m.RegisterMiddleware(MiddlewareRegistration{
				PluginName: name,
				Name:       req.Name,
				Handler:    req.Handler,
			}); err != nil {
				return &pluginv1.BoolResult{Success: false, Error: err.Error()}, nil
			}
			return &pluginv1.BoolResult{Success: true}, nil
		},
		func() *pluginv1.RegisterMiddlewareRequest { return &pluginv1.RegisterMiddlewareRequest{} },
	)
}

// host_enqueue_task — 将任务入队到 asynq
func wrapHostEnqueueTask(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.EnqueueTaskRequest) (*pluginv1.BoolResult, error) {
			if m.hostEnv.Queue == nil {
				return &pluginv1.BoolResult{Success: false, Error: "queue not available"}, nil
			}
			payload := structToMap(req.Payload)
			if err := m.hostEnv.Queue.Enqueue(req.TaskName, payload); err != nil {
				return &pluginv1.BoolResult{Success: false, Error: err.Error()}, nil
			}
			return &pluginv1.BoolResult{Success: true}, nil
		},
		func() *pluginv1.EnqueueTaskRequest { return &pluginv1.EnqueueTaskRequest{} },
	)
}

// ============================================================================
// Helpers
// ============================================================================

var safeConfigKeys = map[string]bool{
	"Site.SiteName": true, "Site.SiteDesc": true, "Site.SiteLogo": true,
	"Site.Host": true, "Currency.Unit": true, "Currency.Symbol": true,
	"Debug": true, "Host": true, "Port": true,
}

func isSafeConfigKey(key string) bool { return safeConfigKeys[key] }

func isURLAllowed(urlStr string) bool {
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	host := u.Hostname()

	if strings.Contains(host, "metadata.google.internal") {
		return false
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return false
		}
	}

	return true
}

func structToMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

func toStringMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// ============================================================================
// Async Runtime Host Functions — goroutine 池
// ============================================================================

// wrapHostAsyncSubmit: guest 调用 host_async_submit(op_type, params) → task_id
func wrapHostAsyncSubmit(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, _ string, req *pluginv1.AsyncSubmitRequest) (*pluginv1.AsyncSubmitResponse, error) {
			// pluginName is captured from the outer closure to tag async tasks
			id := m.submitAsyncTask(pluginName, req.OpType, req.Params)
			return &pluginv1.AsyncSubmitResponse{TaskId: id}, nil
		},
		func() *pluginv1.AsyncSubmitRequest { return &pluginv1.AsyncSubmitRequest{} },
	)
}

// wrapHostAsyncResolve: guest 调用 host_async_resolve(task_id) → result
func wrapHostAsyncResolve(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, req *pluginv1.AsyncResolveRequest) (*pluginv1.AsyncResolveResponse, error) {
			result, errStr, ok := m.resolveAsyncTask(name, req.TaskId)
			return &pluginv1.AsyncResolveResponse{
				TaskId: req.TaskId,
				Done:   ok,
				Result: result,
				Error:  errStr,
			}, nil
		},
		func() *pluginv1.AsyncResolveRequest { return &pluginv1.AsyncResolveRequest{} },
	)
}

// wrapHostAsyncWaitAny: guest 调用 host_async_wait_any() → task_id of completed task
func wrapHostAsyncWaitAny(pluginName string, m *Manager) api.GoModuleFunc {
	return MakeHostFunc(pluginName,
		func(ctx context.Context, name string, _ *pluginv1.AsyncWaitAnyRequest) (*pluginv1.AsyncWaitAnyResponse, error) {
			id := m.waitAnyAsyncTask(name)
			return &pluginv1.AsyncWaitAnyResponse{TaskId: id}, nil
		},
		func() *pluginv1.AsyncWaitAnyRequest { return &pluginv1.AsyncWaitAnyRequest{} },
	)
}

// doHTTP performs an actual outbound HTTP request (used by async HTTP tasks)
func doHTTP(ctx context.Context, preq *pluginv1.HttpRequestRequest) (*pluginv1.HttpRequestResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				if strings.Contains(host, "metadata.google.internal") {
					return nil, fmt.Errorf("metadata server access denied")
				}
				ips, err := net.LookupIP(host)
				if err != nil {
					return nil, err
				}
				var targetIP net.IP
				for _, ip := range ips {
					if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
						continue // skip private IPs
					}
					if targetIP == nil {
						targetIP = ip
					}
				}
				if targetIP == nil {
					return nil, fmt.Errorf("no safe public IP found for host %s", host)
				}
				return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(targetIP.String(), port))
			},
		},
	}
	var reqBody io.Reader
	if len(preq.Body) > 0 {
		reqBody = bytes.NewReader(preq.Body)
	}
	req, err := http.NewRequestWithContext(ctx, preq.Method, preq.Url, reqBody)
	if err != nil {
		return nil, err
	}
	for k, v := range preq.Headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // 限制最大读取 10MB
	respHeaders := make(map[string]string)
	for k, vs := range resp.Header {
		if len(vs) > 0 {
			respHeaders[k] = vs[0]
		}
	}
	return &pluginv1.HttpRequestResponse{
		Status: int32(resp.StatusCode), Body: body, Headers: respHeaders,
	}, nil
}
