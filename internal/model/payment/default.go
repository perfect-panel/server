package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customPaymentModel)(nil)
var (
	cachePaymentIdPrefix    = "cache:payment:id:"
	cachePaymentTokenPrefix = "cache:payment:token:"
)

type (
	Model interface {
		paymentModel
		customPaymentLogicModel
	}
	paymentModel interface {
		Insert(ctx context.Context, data *Payment, tx ...*gorm.DB) error
		FindOne(ctx context.Context, id int64) (*Payment, error)
		Update(ctx context.Context, data *Payment, tx ...*gorm.DB) error
		Delete(ctx context.Context, id int64, tx ...*gorm.DB) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customPaymentModel struct {
		*defaultPaymentModel
	}
	defaultPaymentModel struct {
		cache.CachedConn
		table string
	}
)

func newPaymentModel(db *gorm.DB, c *redis.Client) *defaultPaymentModel {
	return &defaultPaymentModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`Payment`",
	}
}

//nolint:unused
func (m *defaultPaymentModel) batchGetCacheKeys(Payments ...*Payment) []string {
	var keys []string
	for _, payment := range Payments {
		keys = append(keys, m.getCacheKeys(payment)...)
	}
	return keys

}
func (m *defaultPaymentModel) getCacheKeys(data *Payment) []string {
	if data == nil {
		return []string{}
	}
	paymentIdKey := fmt.Sprintf("%s%v", cachePaymentIdPrefix, data.Id)
	paymentNameKey := fmt.Sprintf("%s%v", cachePaymentTokenPrefix, data.Token)
	cacheKeys := []string{
		paymentIdKey,
		paymentNameKey,
	}
	return cacheKeys
}

func (m *defaultPaymentModel) Insert(ctx context.Context, data *Payment, tx ...*gorm.DB) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultPaymentModel) FindOne(ctx context.Context, id int64) (*Payment, error) {
	PaymentIdKey := fmt.Sprintf("%s%v", cachePaymentIdPrefix, id)
	var resp Payment
	err := m.QueryCtx(ctx, &resp, PaymentIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Payment{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultPaymentModel) Update(ctx context.Context, data *Payment, tx ...*gorm.DB) error {
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

func (m *defaultPaymentModel) Delete(ctx context.Context, id int64, tx ...*gorm.DB) error {
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
		return conn.Delete(&Payment{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultPaymentModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
