package ads

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customAdsModel)(nil)
var (
	cacheAdsIdPrefix = "cache:ads:id:"
)

type (
	Model interface {
		adsModel
		customAdsLogicModel
	}
	adsModel interface {
		Insert(ctx context.Context, data *Ads) error
		FindOne(ctx context.Context, id int64) (*Ads, error)
		Update(ctx context.Context, data *Ads) error
		Delete(ctx context.Context, id int64) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customAdsModel struct {
		*defaultAdsModel
	}
	defaultAdsModel struct {
		cache.CachedConn
		table string
	}
)

func newAdsModel(db *gorm.DB, c *redis.Client) *defaultAdsModel {
	return &defaultAdsModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`ads`",
	}
}

//nolint:unused
func (m *defaultAdsModel) batchGetCacheKeys(ads ...*Ads) []string {
	var keys []string
	for _, ad := range ads {
		keys = append(keys, m.getCacheKeys(ad)...)
	}
	return keys

}
func (m *defaultAdsModel) getCacheKeys(data *Ads) []string {
	if data == nil {
		return []string{}
	}
	adsIdKey := fmt.Sprintf("%s%v", cacheAdsIdPrefix, data.Id)
	cacheKeys := []string{
		adsIdKey,
	}
	return cacheKeys
}

func (m *defaultAdsModel) Insert(ctx context.Context, data *Ads) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultAdsModel) FindOne(ctx context.Context, id int64) (*Ads, error) {
	AdsIdKey := fmt.Sprintf("%s%v", cacheAdsIdPrefix, id)
	var resp Ads
	err := m.QueryCtx(ctx, &resp, AdsIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Ads{}).Where("`id` = ?", id).First(&resp).Error
	})
	return &resp, err
}

func (m *defaultAdsModel) Update(ctx context.Context, data *Ads) error {
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

func (m *defaultAdsModel) Delete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Delete(&Ads{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultAdsModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
