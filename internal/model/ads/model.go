package ads

import (
	"context"

	"github.com/perfect-panel/server/pkg/orm"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type customAdsLogicModel interface {
	GetAdsListByPage(ctx context.Context, page, size int, filter Filter) (int64, []*Ads, error)
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customAdsModel{
		defaultAdsModel: newAdsModel(conn, c),
	}
}

type Filter struct {
	Status *int
	Search string
}

// GetAdsListByPage  get ads list by page
func (m *customAdsModel) GetAdsListByPage(ctx context.Context, page, size int, filter Filter) (int64, []*Ads, error) {
	var list []*Ads
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.Model(&Ads{})
		if filter.Status != nil {
			conn = conn.Where("status = ?", *filter.Status)
		}
		if filter.Search != "" {
			conn = conn.Scopes(orm.ContainsLike([]string{"title", "content"}, filter.Search))
		}
		return conn.Count(&total).Offset((page - 1) * size).Limit(size).Find(v).Error
	})
	return total, list, err
}
