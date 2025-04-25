package subscribeType

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customSubscribeTypeModel)(nil)
var (
	cacheSubscribeTypeIdPrefix = "cache:subscribeType:id:"
)

type (
	Model interface {
		subscribeTypeModel
		customSubscribeTypeLogicModel
	}
	subscribeTypeModel interface {
		Insert(ctx context.Context, data *SubscribeType) error
		FindOne(ctx context.Context, id int64) (*SubscribeType, error)
		Update(ctx context.Context, data *SubscribeType) error
		Delete(ctx context.Context, id int64) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customSubscribeTypeModel struct {
		*defaultSubscribeTypeModel
	}
	defaultSubscribeTypeModel struct {
		cache.CachedConn
		table string
	}
)

func newSubscribeTypeModel(db *gorm.DB, c *redis.Client) *defaultSubscribeTypeModel {
	return &defaultSubscribeTypeModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`SubscribeType`",
	}
}

//nolint:unused
func (m *defaultSubscribeTypeModel) batchGetCacheKeys(SubscribeTypes ...*SubscribeType) []string {
	var keys []string
	for _, subscribeType := range SubscribeTypes {
		keys = append(keys, m.getCacheKeys(subscribeType)...)
	}
	return keys

}
func (m *defaultSubscribeTypeModel) getCacheKeys(data *SubscribeType) []string {
	if data == nil {
		return []string{}
	}
	SubscribeTypeIdKey := fmt.Sprintf("%s%v", cacheSubscribeTypeIdPrefix, data.Id)
	cacheKeys := []string{
		SubscribeTypeIdKey,
	}
	return cacheKeys
}

func (m *defaultSubscribeTypeModel) Insert(ctx context.Context, data *SubscribeType) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultSubscribeTypeModel) FindOne(ctx context.Context, id int64) (*SubscribeType, error) {
	SubscribeTypeIdKey := fmt.Sprintf("%s%v", cacheSubscribeTypeIdPrefix, id)
	var resp SubscribeType
	err := m.QueryCtx(ctx, &resp, SubscribeTypeIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&SubscribeType{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultSubscribeTypeModel) Update(ctx context.Context, data *SubscribeType) error {
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

func (m *defaultSubscribeTypeModel) Delete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Delete(&SubscribeType{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultSubscribeTypeModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
