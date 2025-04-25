package cache

import (
	"context"
	"time"
)

type (
	Cache interface {
		// Del deletes cached values with keys.
		Del(keys ...string) error
		// DelCtx deletes cached values with keys.
		DelCtx(ctx context.Context, keys ...string) error
		// Get gets the cache with key and fills into v.
		Get(key string, val any) error
		// GetCtx gets the cache with key and fills into v.
		GetCtx(ctx context.Context, key string, val any) error
		// Set sets the cache with key and value.
		Set(key string, val any) error
		// SetCtx sets the cache with key and value.
		SetCtx(ctx context.Context, key string, val any) error
		// SetWithExpire sets the cache with key and v, using given expire.
		SetWithExpire(key string, val any, expire time.Duration) error
		// SetWithExpireCtx sets the cache with key and v, using given expire.
		SetWithExpireCtx(ctx context.Context, key string, val any, expire time.Duration) error
	}
)
