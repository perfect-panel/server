package coupon

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customCouponModel)(nil)
var (
	cacheCouponIdPrefix   = "cache:coupon:id:"
	cacheCouponCodePrefix = "cache:coupon:code:"
)

type (
	Model interface {
		couponModel
		customCouponLogicModel
	}
	couponModel interface {
		Insert(ctx context.Context, data *Coupon) error
		FindOne(ctx context.Context, id int64) (*Coupon, error)
		FindOneByCode(ctx context.Context, code string) (*Coupon, error)
		Update(ctx context.Context, data *Coupon) error
		Delete(ctx context.Context, id int64) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customCouponModel struct {
		*defaultCouponModel
	}
	defaultCouponModel struct {
		cache.CachedConn
		table string
	}
)

func newCouponModel(db *gorm.DB, c *redis.Client) *defaultCouponModel {
	return &defaultCouponModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`coupon`",
	}
}

//nolint:unused
func (m *defaultCouponModel) batchGetCacheKeys(Coupons ...*Coupon) []string {
	var keys []string
	for _, coupon := range Coupons {
		keys = append(keys, m.getCacheKeys(coupon)...)
	}
	return keys

}
func (m *defaultCouponModel) getCacheKeys(data *Coupon) []string {
	if data == nil {
		return []string{}
	}
	couponIdKey := fmt.Sprintf("%s%v", cacheCouponIdPrefix, data.Id)
	couponCodeKey := fmt.Sprintf("%s%v", cacheCouponCodePrefix, data.Code)
	cacheKeys := []string{
		couponIdKey,
		couponCodeKey,
	}
	return cacheKeys
}

func (m *defaultCouponModel) Insert(ctx context.Context, data *Coupon) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultCouponModel) FindOne(ctx context.Context, id int64) (*Coupon, error) {
	CouponIdKey := fmt.Sprintf("%s%v", cacheCouponIdPrefix, id)
	var resp Coupon
	err := m.QueryCtx(ctx, &resp, CouponIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Coupon{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultCouponModel) FindOneByCode(ctx context.Context, code string) (*Coupon, error) {
	CouponCodeKey := fmt.Sprintf("%s%v", cacheCouponCodePrefix, code)
	var resp Coupon
	err := m.QueryCtx(ctx, &resp, CouponCodeKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Coupon{}).Where("`code` = ?", code).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultCouponModel) Update(ctx context.Context, data *Coupon) error {
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

func (m *defaultCouponModel) Delete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Delete(&Coupon{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultCouponModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
