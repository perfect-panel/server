package coupon

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type customCouponLogicModel interface {
	UpdateCount(ctx context.Context, code string) error
	QueryCouponListByPage(ctx context.Context, page, size int, subscribe int64, search string) (total int64, list []*Coupon, err error)
	BatchDelete(ctx context.Context, ids []int64) error
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customCouponModel{
		defaultCouponModel: newCouponModel(conn, c),
	}
}

// QueryCouponListByPage query coupon list by page
func (m *customCouponModel) QueryCouponListByPage(ctx context.Context, page, size int, subscribe int64, search string) (total int64, list []*Coupon, err error) {
	err = m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		db := conn.Model(&Coupon{})
		if subscribe != 0 {
			db = db.Where("FIND_IN_SET(?, subscribe)", subscribe)
		}
		if search != "" {
			db = db.Where("name like ? or code like ?", "%"+search+"%", "%"+search+"%")
		}
		return db.Count(&total).Limit(size).Offset((page - 1) * size).Find(v).Error
	})
	return total, list, err
}

func (m *customCouponModel) BatchDelete(ctx context.Context, ids []int64) error {
	var err error
	for _, id := range ids {
		if err = m.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (m *customCouponModel) UpdateCount(ctx context.Context, code string) error {
	data, err := m.FindOneByCode(ctx, code)
	if err != nil {
		return err
	}
	data.UsedCount++
	return m.Update(ctx, data)
}
