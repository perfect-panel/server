package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// RedisAdapter 将 *redis.Client 适配为 RedisClient 接口
type RedisAdapter struct {
	client *redis.Client
}

// NewRedisAdapter 创建 Redis 适配器
func NewRedisAdapter(client *redis.Client) *RedisAdapter {
	return &RedisAdapter{client: client}
}

func (a *RedisAdapter) Get(key string) (string, error) {
	ctx := context.Background()
	val, err := a.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, err
}

func (a *RedisAdapter) Set(key string, value string, ttlSeconds int64) error {
	ctx := context.Background()
	return a.client.Set(ctx, key, value, time.Duration(ttlSeconds)*time.Second).Err()
}

func (a *RedisAdapter) Del(keys ...string) error {
	ctx := context.Background()
	return a.client.Del(ctx, keys...).Err()
}

// QueueAdapter 将 *asynq.Client 适配为 QueueClient 接口
type QueueAdapter struct {
	client *asynq.Client
}

// NewQueueAdapter 创建 Queue 适配器
func NewQueueAdapter(client *asynq.Client) *QueueAdapter {
	return &QueueAdapter{client: client}
}

// Enqueue 将任务入队
func (a *QueueAdapter) Enqueue(taskName string, payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	task := asynq.NewTask(taskName, data)
	_, err = a.client.Enqueue(task)
	return err
}
