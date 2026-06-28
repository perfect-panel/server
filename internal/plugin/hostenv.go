package plugin

import "github.com/perfect-panel/server/internal/config"

// HostEnv 封装宿主提供给插件的所有依赖
// 在 cmd/run.go 中由 ServiceContext 构建
type HostEnv struct {
	Config config.Config
	Redis  RedisClient
	Store  StoreClient
	Queue  QueueClient
}

// RedisClient 是插件访问 Redis 的接口
type RedisClient interface {
	Get(key string) (string, error)
	Set(key string, value string, ttlSeconds int64) error
	Del(keys ...string) error
}

// StoreClient 是插件访问数据库的接口
type StoreClient interface {
	Query(model string, operation string, conditions map[string]interface{}, fields []string, limit, offset int32) ([]map[string]interface{}, int64, error)
}

// QueueClient 是插件访问任务队列的接口
type QueueClient interface {
	Enqueue(taskName string, payload map[string]interface{}) error
}
