package node

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type (
	customCacheLogicModel interface {
		StatusCache(ctx context.Context, serverId int64) (Status, error)
		UpdateStatusCache(ctx context.Context, serverId int64, status *Status) error
		OnlineUserSubscribe(ctx context.Context, serverId int64, protocol string) (OnlineUserSubscribe, error)
		UpdateOnlineUserSubscribe(ctx context.Context, serverId int64, protocol string, subscribe OnlineUserSubscribe) error
		OnlineUserSubscribeGlobal(ctx context.Context) (int64, error)
		UpdateOnlineUserSubscribeGlobal(ctx context.Context, subscribe OnlineUserSubscribe) error
	}

	Status struct {
		Cpu       float64 `json:"cpu"`
		Mem       float64 `json:"mem"`
		Disk      float64 `json:"disk"`
		UpdatedAt int64   `json:"updated_at"`
	}

	OnlineUserSubscribe map[int64][]string
)

// Marshal  to json string
func (s *Status) Marshal() string {
	type Alias Status
	data, _ := json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
	return string(data)
}

// Unmarshal from json string
func (s *Status) Unmarshal(data string) error {
	type Alias Status
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal([]byte(data), &aux)
}

const (
	Expiry                                = 300 * time.Second              // Cache expiry time in seconds
	StatusCacheKey                        = "node:status:%d"               // Node status cache key format (Server ID and protocol) Example: node:status:1:shadowsocks
	OnlineUserCacheKeyWithSubscribe       = "node:online:subscribe:%d:%s"  // Online user subscribe cache key format (Server ID and protocol) Example: node:online:subscribe:1:shadowsocks
	OnlineUserSubscribeCacheKeyWithGlobal = "node:online:subscribe:global" // Online user global subscribe cache key
)

// UpdateStatusCache Update server status to cache
func (m *customServerModel) UpdateStatusCache(ctx context.Context, serverId int64, status *Status) error {
	key := fmt.Sprintf(StatusCacheKey, serverId)
	return m.Cache.Set(ctx, key, status.Marshal(), Expiry).Err()

}

// DeleteStatusCache Delete server status from cache
func (m *customServerModel) DeleteStatusCache(ctx context.Context, serverId int64) error {
	key := fmt.Sprintf(StatusCacheKey, serverId)
	return m.Cache.Del(ctx, key).Err()
}

// StatusCache Get server status from cache
func (m *customServerModel) StatusCache(ctx context.Context, serverId int64) (Status, error) {
	var status Status
	key := fmt.Sprintf(StatusCacheKey, serverId)

	result, err := m.Cache.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return status, nil
		}
		return status, err
	}
	if result == "" {
		return status, nil
	}
	err = status.Unmarshal(result)
	return status, err
}

// OnlineUserSubscribe Get online user subscribe
func (m *customServerModel) OnlineUserSubscribe(ctx context.Context, serverId int64, protocol string) (OnlineUserSubscribe, error) {
	key := fmt.Sprintf(OnlineUserCacheKeyWithSubscribe, serverId, protocol)
	result, err := m.Cache.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return OnlineUserSubscribe{}, nil
		}
		return nil, err
	}
	if result == "" {
		return OnlineUserSubscribe{}, nil
	}
	var subscribe OnlineUserSubscribe
	err = json.Unmarshal([]byte(result), &subscribe)
	return subscribe, err
}

// UpdateOnlineUserSubscribe Update online user subscribe
func (m *customServerModel) UpdateOnlineUserSubscribe(ctx context.Context, serverId int64, protocol string, subscribe OnlineUserSubscribe) error {
	key := fmt.Sprintf(OnlineUserCacheKeyWithSubscribe, serverId, protocol)
	data, err := json.Marshal(subscribe)
	if err != nil {
		return err
	}
	return m.Cache.Set(ctx, key, data, Expiry).Err()
}

// DeleteOnlineUserSubscribe Delete online user subscribe
func (m *customServerModel) DeleteOnlineUserSubscribe(ctx context.Context, serverId int64, protocol string) error {
	key := fmt.Sprintf(OnlineUserCacheKeyWithSubscribe, serverId, protocol)
	return m.Cache.Del(ctx, key).Err()
}

// OnlineUserSubscribeGlobal Get global online user subscribe count
func (m *customServerModel) OnlineUserSubscribeGlobal(ctx context.Context) (int64, error) {
	now := time.Now().Unix()
	// Clear expired data
	if err := m.Cache.ZRemRangeByScore(ctx, OnlineUserSubscribeCacheKeyWithGlobal, "-inf", fmt.Sprintf("%d", now)).Err(); err != nil {
		return 0, err
	}
	return m.Cache.ZCard(ctx, OnlineUserSubscribeCacheKeyWithGlobal).Result()
}

// UpdateOnlineUserSubscribeGlobal Update global online user subscribe count
func (m *customServerModel) UpdateOnlineUserSubscribeGlobal(ctx context.Context, subscribe OnlineUserSubscribe) error {
	now := time.Now()
	expireTime := now.Add(5 * time.Minute).Unix() // set expire time 5 minutes later

	pipe := m.Cache.Pipeline()

	// Clear expired data
	pipe.ZRemRangeByScore(ctx, OnlineUserSubscribeCacheKeyWithGlobal, "-inf", fmt.Sprintf("%d", now.Unix()))
	// Add or update each subscribe with new expire time
	for sub := range subscribe {
		// Use ZAdd to add or update the member with new score (expire time)
		pipe.ZAdd(ctx, OnlineUserSubscribeCacheKeyWithGlobal, redis.Z{
			Score:  float64(expireTime),
			Member: sub,
		})
	}

	_, err := pipe.Exec(ctx)
	return err
}

// DeleteOnlineUserSubscribeGlobal Delete global online user subscribe count
func (m *customServerModel) DeleteOnlineUserSubscribeGlobal(ctx context.Context) error {
	return m.Cache.Del(ctx, OnlineUserSubscribeCacheKeyWithGlobal).Err()
}
