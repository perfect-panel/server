package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ErrNotFound = redis.Nil

type (
	// ExecCtxFn defines the sql exec method.
	ExecCtxFn func(conn *gorm.DB) error
	// IndexQueryCtxFn defines the query method that based on unique indexes.
	IndexQueryCtxFn func(conn *gorm.DB, v interface{}) (interface{}, error)
	// PrimaryQueryCtxFn defines the query method that based on primary keys.
	PrimaryQueryCtxFn func(conn *gorm.DB, v, primary interface{}) error
	// QueryCtxFn defines the query method.
	QueryCtxFn func(conn *gorm.DB, v interface{}) error

	CachedConn struct {
		db    *gorm.DB
		cache *redis.Client
	}
)

// NewConn returns a CachedConn with a redis cluster cache.
func NewConn(db *gorm.DB, c *redis.Client) CachedConn {
	return CachedConn{
		db:    db,
		cache: c,
	}
}

// DelCache deletes cache with keys.
func (cc CachedConn) DelCache(keys ...string) error {
	return cc.cache.Del(context.Background(), keys...).Err()
}

// DelCacheCtx deletes cache with keys.
func (cc CachedConn) DelCacheCtx(ctx context.Context, keys ...string) error {
	return cc.cache.Del(ctx, keys...).Err()
}

// GetCache unmarshals cache with given key into v.
func (cc CachedConn) GetCache(key string, v interface{}) error {
	// query redis key
	val, err := cc.cache.Get(context.Background(), key).Result()
	if err != nil {
		return err
	}
	// unmarshal value
	return json.Unmarshal([]byte(val), v)
}

// SetCache sets cache with key and v.
func (cc CachedConn) SetCache(key string, v interface{}) error {
	// marshal value
	val, err := json.Marshal(v)
	if err != nil {
		return err
	}
	// set redis key
	return cc.cache.Set(context.Background(), key, val, 0).Err()
}

// ExecCtx runs given exec on given keys, and returns execution result.
func (cc CachedConn) ExecCtx(ctx context.Context, execCtx ExecCtxFn, keys ...string) error {
	err := execCtx(cc.db.WithContext(ctx))
	if err != nil {
		return err
	}
	if err := cc.DelCacheCtx(ctx, keys...); err != nil {
		return err
	}
	return nil
}

// ExecNoCache runs exec with given sql statement, without affecting cache.
func (cc CachedConn) ExecNoCache(exec ExecCtxFn) error {
	return cc.ExecNoCacheCtx(context.Background(), exec)
}

// ExecNoCacheCtx runs exec with given sql statement, without affecting cache.
func (cc CachedConn) ExecNoCacheCtx(ctx context.Context, execCtx ExecCtxFn) (err error) {
	return execCtx(cc.db.WithContext(ctx))
}

func (cc CachedConn) QueryCtx(ctx context.Context, v interface{}, key string, query QueryCtxFn) (err error) {
	err = cc.GetCache(key, v)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			err = query(cc.db.WithContext(ctx), v)
			if err != nil {
				return err
			}
			return cc.SetCache(key, v)
		}
	}
	return
}

// QueryNoCacheCtx runs query with given sql statement, without affecting cache.
func (cc CachedConn) QueryNoCacheCtx(ctx context.Context, v interface{}, query QueryCtxFn) (err error) {
	return query(cc.db.WithContext(ctx), v)
}

// TransactCtx runs given fn in transaction mode.
func (cc CachedConn) TransactCtx(ctx context.Context, fn func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return cc.db.WithContext(ctx).Transaction(fn, opts...)
}

// Transact runs given fn in transaction mode.
func (cc CachedConn) Transact(fn func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return cc.TransactCtx(context.Background(), fn, opts...)
}
