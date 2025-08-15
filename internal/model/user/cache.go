package user

import (
	"context"
	"fmt"

	"github.com/perfect-panel/server/pkg/logger"
)

type CacheKeyGenerator interface {
	GetCacheKeys() []string
}

type CacheManager interface {
	ClearCache(ctx context.Context, keys ...string) error
	ClearModelCache(ctx context.Context, models ...CacheKeyGenerator) error
}

type UserCacheManager struct {
	model *defaultUserModel
}

func NewUserCacheManager(model *defaultUserModel) *UserCacheManager {
	return &UserCacheManager{
		model: model,
	}
}

func (c *UserCacheManager) ClearCache(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.model.CachedConn.DelCacheCtx(ctx, keys...)
}

func (c *UserCacheManager) ClearModelCache(ctx context.Context, models ...CacheKeyGenerator) error {
	var allKeys []string
	for _, model := range models {
		if model != nil {
			allKeys = append(allKeys, model.GetCacheKeys()...)
		}
	}
	return c.ClearCache(ctx, allKeys...)
}

func (u *User) GetCacheKeys() []string {
	if u == nil {
		return []string{}
	}
	keys := []string{
		fmt.Sprintf("%s%d", cacheUserIdPrefix, u.Id),
	}

	for _, auth := range u.AuthMethods {
		if auth.AuthType == "email" {
			keys = append(keys, fmt.Sprintf("%s%s", cacheUserEmailPrefix, auth.AuthIdentifier))
			break
		}
	}
	return keys
}

func (s *Subscribe) GetCacheKeys() []string {
	if s == nil {
		return []string{}
	}
	keys := make([]string, 0)

	if s.Token != "" {
		keys = append(keys, fmt.Sprintf("%s%s", cacheUserSubscribeTokenPrefix, s.Token))
	}
	if s.UserId != 0 {
		keys = append(keys, fmt.Sprintf("%s%d", cacheUserSubscribeUserPrefix, s.UserId))
	}
	if s.Id != 0 {
		keys = append(keys, fmt.Sprintf("%s%d", cacheUserSubscribeIdPrefix, s.Id))
	}
	return keys
}

func (s *Subscribe) GetExtendedCacheKeys(model *defaultUserModel) []string {
	keys := s.GetCacheKeys()

	if s.SubscribeId != 0 && model != nil {
		serverKeys := model.getServerRelatedCacheKeys(s.SubscribeId)
		keys = append(keys, serverKeys...)
	}

	return keys
}

func (d *Device) GetCacheKeys() []string {
	if d == nil {
		return []string{}
	}
	keys := []string{}

	if d.Id != 0 {
		keys = append(keys, fmt.Sprintf("%s%d", cacheUserDeviceIdPrefix, d.Id))
	}
	if d.Identifier != "" {
		keys = append(keys, fmt.Sprintf("%s%s", cacheUserDeviceNumberPrefix, d.Identifier))
	}
	return keys
}

func (a *AuthMethods) GetCacheKeys() []string {
	if a == nil {
		return []string{}
	}
	keys := []string{}

	if a.UserId != 0 {
		keys = append(keys, fmt.Sprintf("%s%d", cacheUserIdPrefix, a.UserId))
	}
	if a.AuthType == "email" && a.AuthIdentifier != "" {
		keys = append(keys, fmt.Sprintf("%s%s", cacheUserEmailPrefix, a.AuthIdentifier))
	}
	return keys
}

func (m *defaultUserModel) GetCacheManager() *UserCacheManager {
	return NewUserCacheManager(m)
}

func (m *defaultUserModel) getServerRelatedCacheKeys(subscribeId int64) []string {
	// 这里复用了 model.go 中的逻辑，但简化了实现
	keys := []string{}

	if subscribeId == 0 {
		return keys
	}

	// 这里需要从 getSubscribeCacheKey 方法中提取服务器相关的逻辑
	// 为了避免重复查询，我们可以在需要时才获取
	// 或者可以将这个逻辑移到一个统一的地方

	return keys
}

func (m *defaultUserModel) ClearUserCache(ctx context.Context, users ...*User) error {
	cacheManager := m.GetCacheManager()
	models := make([]CacheKeyGenerator, len(users))
	for i, user := range users {
		models[i] = user
	}
	return cacheManager.ClearModelCache(ctx, models...)
}

func (m *defaultUserModel) ClearSubscribeCacheByModels(ctx context.Context, subscribes ...*Subscribe) error {
	cacheManager := m.GetCacheManager()
	models := make([]CacheKeyGenerator, len(subscribes))
	for i, subscribe := range subscribes {
		models[i] = subscribe
	}
	return cacheManager.ClearModelCache(ctx, models...)
}

func (m *defaultUserModel) ClearDeviceCache(ctx context.Context, devices ...*Device) error {
	cacheManager := m.GetCacheManager()
	models := make([]CacheKeyGenerator, len(devices))
	for i, device := range devices {
		models[i] = device
	}
	return cacheManager.ClearModelCache(ctx, models...)
}

func (m *defaultUserModel) ClearAuthMethodCache(ctx context.Context, authMethods ...*AuthMethods) error {
	cacheManager := m.GetCacheManager()
	models := make([]CacheKeyGenerator, len(authMethods))
	for i, auth := range authMethods {
		models[i] = auth
	}
	return cacheManager.ClearModelCache(ctx, models...)
}

func (m *defaultUserModel) BatchClearRelatedCache(ctx context.Context, user *User) error {
	if user == nil {
		return nil
	}

	cacheManager := m.GetCacheManager()

	var allModels []CacheKeyGenerator
	allModels = append(allModels, user)

	for _, auth := range user.AuthMethods {
		allModels = append(allModels, &auth)
	}

	for _, device := range user.UserDevices {
		allModels = append(allModels, &device)
	}

	subscribes, err := m.QueryUserSubscribe(ctx, user.Id)
	if err != nil {
		logger.Errorf("failed to query user subscribes for cache clearing: %v", err)
	} else {
		for _, sub := range subscribes {
			subModel := &Subscribe{
				Id:          sub.Id,
				UserId:      sub.UserId,
				Token:       sub.Token,
				SubscribeId: sub.SubscribeId,
			}
			allModels = append(allModels, subModel)
		}
	}

	return cacheManager.ClearModelCache(ctx, allModels...)
}

func (m *defaultUserModel) CacheInvalidationHandler(ctx context.Context, operation string, modelType string, model interface{}) error {
	switch operation {
	case "create", "update", "delete":
		switch modelType {
		case "user":
			if user, ok := model.(*User); ok {
				return m.BatchClearRelatedCache(ctx, user)
			}
		case "subscribe":
			if subscribe, ok := model.(*Subscribe); ok {
				return m.ClearSubscribeCacheByModels(ctx, subscribe)
			}
		case "device":
			if device, ok := model.(*Device); ok {
				return m.ClearDeviceCache(ctx, device)
			}
		case "authmethod":
			if authMethod, ok := model.(*AuthMethods); ok {
				return m.ClearAuthMethodCache(ctx, authMethod)
			}
		}
	}
	return nil
}

func (m *customUserModel) GetRelatedCacheKeys(ctx context.Context, modelType string, modelId int64) ([]string, error) {
	var keys []string

	switch modelType {
	case "user":
		user, err := m.FindOne(ctx, modelId)
		if err != nil {
			return nil, err
		}
		keys = append(keys, user.GetCacheKeys()...)

		auths, err := m.FindUserAuthMethods(ctx, modelId)
		if err == nil {
			for _, auth := range auths {
				keys = append(keys, auth.GetCacheKeys()...)
			}
		}

		subscribes, err := m.QueryUserSubscribe(ctx, modelId)
		if err == nil {
			for _, sub := range subscribes {
				subModel := &Subscribe{
					Id:          sub.Id,
					UserId:      sub.UserId,
					Token:       sub.Token,
					SubscribeId: sub.SubscribeId,
				}
				keys = append(keys, subModel.GetCacheKeys()...)
			}
		}

	case "subscribe":
		subscribe, err := m.FindOneSubscribe(ctx, modelId)
		if err != nil {
			return nil, err
		}
		keys = append(keys, subscribe.GetCacheKeys()...)

	case "device":
		device, err := m.FindOneDevice(ctx, modelId)
		if err != nil {
			return nil, err
		}
		keys = append(keys, device.GetCacheKeys()...)
	}

	return keys, nil
}
