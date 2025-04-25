package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customAuthModel)(nil)
var (
	cacheAuthIdPrefix     = "cache:auth:id:"
	cacheAuthMethodPrefix = "cache:auth:method:"
)

type (
	Model interface {
		authModel
		customAuthLogicModel
	}
	authModel interface {
		Insert(ctx context.Context, data *Auth) error
		FindOne(ctx context.Context, id int64) (*Auth, error)
		Update(ctx context.Context, data *Auth) error
		Delete(ctx context.Context, id int64) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customAuthModel struct {
		*defaultAuthModel
	}
	defaultAuthModel struct {
		cache.CachedConn
		table string
	}
)

func newAuthModel(db *gorm.DB, c *redis.Client) *defaultAuthModel {
	return &defaultAuthModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`auth_config`",
	}
}

//nolint:unused
func (m *defaultAuthModel) batchGetCacheKeys(Auths ...*Auth) []string {
	var keys []string
	for _, auth := range Auths {
		keys = append(keys, m.getCacheKeys(auth)...)
	}
	return keys

}
func (m *defaultAuthModel) getCacheKeys(data *Auth) []string {
	if data == nil {
		return []string{}
	}
	authIdKey := fmt.Sprintf("%s%v", cacheAuthIdPrefix, data.Id)
	platformKey := fmt.Sprintf("%s%s", cacheAuthMethodPrefix, data.Method)
	cacheKeys := []string{
		authIdKey,
		platformKey,
	}
	return cacheKeys
}

func (m *defaultAuthModel) Insert(ctx context.Context, data *Auth) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultAuthModel) FindOne(ctx context.Context, id int64) (*Auth, error) {
	AuthIdKey := fmt.Sprintf("%s%v", cacheAuthIdPrefix, id)
	var resp Auth
	err := m.QueryCtx(ctx, &resp, AuthIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Auth{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultAuthModel) Update(ctx context.Context, data *Auth) error {
	old, err := m.FindOne(ctx, data.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Save(data).Error
	}, m.getCacheKeys(old)...)
	return err
}

func (m *defaultAuthModel) Delete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Delete(&Auth{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultAuthModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
