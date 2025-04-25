package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customOrderModel)(nil)
var (
	cacheOrderIdPrefix = "cache:order:id:"
	cacheOrderNoPrefix = "cache:order:no:"
)

type (
	Model interface {
		orderModel
		customOrderLogicModel
	}
	orderModel interface {
		Insert(ctx context.Context, data *Order, tx ...*gorm.DB) error
		FindOne(ctx context.Context, id int64) (*Order, error)
		FindOneByOrderNo(ctx context.Context, orderNo string) (*Order, error)
		Update(ctx context.Context, data *Order, tx ...*gorm.DB) error
		Delete(ctx context.Context, id int64, tx ...*gorm.DB) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customOrderModel struct {
		*defaultOrderModel
	}
	defaultOrderModel struct {
		cache.CachedConn
		table string
	}
)

func newOrderModel(db *gorm.DB, c *redis.Client) *defaultOrderModel {
	return &defaultOrderModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`order`",
	}
}

//nolint:unused
func (m *defaultOrderModel) batchGetCacheKeys(Orders ...*Order) []string {
	var keys []string
	for _, order := range Orders {
		keys = append(keys, m.getCacheKeys(order)...)
	}
	return keys

}
func (m *defaultOrderModel) getCacheKeys(data *Order) []string {
	if data == nil {
		return []string{}
	}
	orderIdKey := fmt.Sprintf("%s%v", cacheOrderIdPrefix, data.Id)
	orderNoKey := fmt.Sprintf("%s%v", cacheOrderNoPrefix, data.OrderNo)
	cacheKeys := []string{
		orderIdKey,
		orderNoKey,
	}
	return cacheKeys
}

func (m *defaultOrderModel) Insert(ctx context.Context, data *Order, tx ...*gorm.DB) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultOrderModel) FindOne(ctx context.Context, id int64) (*Order, error) {
	OrderIdKey := fmt.Sprintf("%s%v", cacheOrderIdPrefix, id)
	var resp Order
	err := m.QueryCtx(ctx, &resp, OrderIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Order{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) FindOneByOrderNo(ctx context.Context, orderNo string) (*Order, error) {
	OrderNoKey := fmt.Sprintf("%s%v", cacheOrderNoPrefix, orderNo)
	var resp Order
	err := m.QueryCtx(ctx, &resp, OrderNoKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Order{}).Where("`order_no` = ?", orderNo).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultOrderModel) Update(ctx context.Context, data *Order, tx ...*gorm.DB) error {
	old, err := m.FindOne(ctx, data.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Save(data).Error
	}, m.getCacheKeys(old)...)
	return err
}

func (m *defaultOrderModel) Delete(ctx context.Context, id int64, tx ...*gorm.DB) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Delete(&Order{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultOrderModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
