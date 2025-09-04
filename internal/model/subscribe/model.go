package subscribe

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type customSubscribeLogicModel interface {
	QuerySubscribeListByPage(ctx context.Context, page, size int, lang string, search string) (total int64, list []*Subscribe, err error)
	QuerySubscribeList(ctx context.Context) ([]*Subscribe, error)
	QuerySubscribeListByShow(ctx context.Context, lang string) ([]*Subscribe, error)
	QuerySubscribeIdsByNodeIdAndNodeTag(ctx context.Context, node []int64, tags []string) ([]*Subscribe, error)
	QuerySubscribeMinSortByIds(ctx context.Context, ids []int64) (int64, error)
	QuerySubscribeListByIds(ctx context.Context, ids []int64) ([]*Subscribe, error)
	ClearCache(ctx context.Context, id ...int64) error
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customSubscribeModel{
		defaultSubscribeModel: newSubscribeModel(conn, c),
	}
}

// QuerySubscribeListByPage  Get Subscribe List
func (m *customSubscribeModel) QuerySubscribeListByPage(ctx context.Context, page, size int, lang string, search string) (total int64, list []*Subscribe, err error) {
	err = m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		// About to be abandoned
		_ = conn.Model(&Subscribe{}).
			Where("sort = ?", 0).
			Update("sort", gorm.Expr("id"))

		conn = conn.Model(&Subscribe{})
		if lang != "" {
			conn = conn.Where("`language` = ?", lang)
		}
		if search != "" {
			conn = conn.Where("`name` like ? or `description` like ?", "%"+search+"%", "%"+search+"%")
		}
		err = conn.Count(&total).Order("sort ASC").Limit(size).Offset((page - 1) * size).Find(v).Error
		return nil
	})
	return total, list, err
}

// QuerySubscribeList Get Subscribe List
func (m *customSubscribeModel) QuerySubscribeList(ctx context.Context) ([]*Subscribe, error) {
	var list []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.Model(&Subscribe{})
		return conn.Where("`sell` = true").Order("sort ").Find(v).Error
	})
	return list, err
}

func (m *customSubscribeModel) QuerySubscribeIdsByNodeIdAndNodeTag(ctx context.Context, node []int64, tags []string) ([]*Subscribe, error) {
	var data []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		db := conn.Model(&Subscribe{})
		if len(node) > 0 {
			for _, id := range node {
				db = db.Or("FIND_IN_SET(?, nodes)", id)
			}
		}

		if len(tags) > 0 {
			// 拼接多个 tag 条件
			for _, t := range tags {
				db = db.Or("FIND_IN_SET(?, node_tags)", t)
			}
		}

		return db.Find(v).Error
	})
	return data, err
}

// QuerySubscribeListByShow Get Subscribe List By Show
func (m *customSubscribeModel) QuerySubscribeListByShow(ctx context.Context, lang string) ([]*Subscribe, error) {
	var list []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.Model(&Subscribe{})
		return conn.Where("`show` = true AND `language` = ?", lang).Find(v).Error
	})
	return list, err
}

func (m *customSubscribeModel) QuerySubscribeMinSortByIds(ctx context.Context, ids []int64) (int64, error) {
	var minSort int64
	err := m.QueryNoCacheCtx(ctx, &minSort, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("id IN ?", ids).Select("COALESCE(MIN(sort), 0)").Scan(v).Error
	})
	return minSort, err
}

func (m *customSubscribeModel) QuerySubscribeListByIds(ctx context.Context, ids []int64) ([]*Subscribe, error) {
	var list []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("id IN ?", ids).Find(v).Error
	})
	return list, err
}

func (m *customSubscribeModel) ClearCache(ctx context.Context, ids ...int64) error {
	if len(ids) <= 0 {
		return nil
	}

	var cacheKeys []string
	for _, id := range ids {
		data, err := m.FindOne(ctx, id)
		if err != nil {
			return err
		}
		cacheKeys = append(cacheKeys, m.getCacheKeys(data)...)
	}
	return m.CachedConn.DelCacheCtx(ctx, cacheKeys...)
}
