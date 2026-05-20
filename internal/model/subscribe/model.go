package subscribe

import (
	"context"

	"github.com/perfect-panel/server/pkg/orm"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FilterParams struct {
	Page            int      // Page Number
	Size            int      // Page Size
	Ids             []int64  // Subscribe IDs
	Node            []int64  // Node IDs
	Tags            []string // Node Tags
	Show            bool     // Show Portal Page
	Sell            bool     // Sell
	Language        string   // Language
	DefaultLanguage bool     // Default Subscribe Language Data
	Search          string   // Search Keywords
}

func (p *FilterParams) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Size <= 0 {
		p.Size = 10
	}
}

type customSubscribeLogicModel interface {
	FilterList(ctx context.Context, params *FilterParams) (int64, []*Subscribe, error)
	ClearCache(ctx context.Context, id ...int64) error
	QuerySubscribeMinSortByIds(ctx context.Context, ids []int64) (int64, error)
	UpdateSort(ctx context.Context, data []*Subscribe) error
	QueryGroupList(ctx context.Context) (int64, []*Group, error)
	CreateGroup(ctx context.Context, data *Group) error
	UpdateGroup(ctx context.Context, data *Group) error
	DeleteGroup(ctx context.Context, id int64) error
	BatchDeleteGroup(ctx context.Context, ids []int64) error
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customSubscribeModel{
		defaultSubscribeModel: newSubscribeModel(conn, c),
	}
}

func (m *customSubscribeModel) QuerySubscribeMinSortByIds(ctx context.Context, ids []int64) (int64, error) {
	var minSort int64
	err := m.QueryNoCacheCtx(ctx, &minSort, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("id IN ?", ids).Select("COALESCE(MIN(sort), 0)").Scan(v).Error
	})
	return minSort, err
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

func (m *customSubscribeModel) UpdateSort(ctx context.Context, data []*Subscribe) error {
	if len(data) == 0 {
		return nil
	}
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Save(data).Error
	}, m.batchGetCacheKeys(data...)...)
}

func (m *customSubscribeModel) QueryGroupList(ctx context.Context) (int64, []*Group, error) {
	var list []*Group
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Group{}).Count(&total).Find(v).Error
	})
	return total, list, err
}

func (m *customSubscribeModel) CreateGroup(ctx context.Context, data *Group) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return conn.Model(&Group{}).Create(data).Error
	})
}

func (m *customSubscribeModel) UpdateGroup(ctx context.Context, data *Group) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return conn.Model(&Group{}).Where("id = ?", data.Id).Save(data).Error
	})
}

func (m *customSubscribeModel) DeleteGroup(ctx context.Context, id int64) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return conn.Model(&Group{}).Where("id = ?", id).Delete(&Group{}).Error
	})
}

func (m *customSubscribeModel) BatchDeleteGroup(ctx context.Context, ids []int64) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return conn.Model(&Group{}).Where("id IN ?", ids).Delete(&Group{}).Error
	})
}

// FilterList Filter Subscribe List
func (m *customSubscribeModel) FilterList(ctx context.Context, params *FilterParams) (int64, []*Subscribe, error) {
	if params == nil {
		params = &FilterParams{}
	}
	params.Normalize()

	var list []*Subscribe
	var total int64

	// 构建查询函数
	buildQuery := func(conn *gorm.DB, lang string) *gorm.DB {
		query := conn.Model(&Subscribe{})

		if params.Search != "" {
			query = query.Scopes(orm.ContainsLike([]string{"name", "description"}, params.Search))
		}
		if params.Show {
			query = query.Where(clause.Eq{
				Column: clause.Column{Name: "show"},
				Value:  true,
			})
		}
		if params.Sell {
			query = query.Where("sell = true")
		}

		if len(params.Ids) > 0 {
			query = query.Where("id IN ?", params.Ids)
		}
		if len(params.Node) > 0 {
			query = query.Scopes(InSet("nodes", tool.Int64SliceToStringSlice(params.Node)))
		}

		if len(params.Tags) > 0 {
			query = query.Scopes(InSet("node_tags", params.Tags))
		}
		if lang != "" {
			query = query.Where("language = ?", lang)
		} else if params.DefaultLanguage {
			query = query.Where("language = ''")
		}

		return query
	}

	// 查询数据
	queryFunc := func(lang string) error {
		return m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
			query := buildQuery(conn, lang)
			if err := query.Count(&total).Error; err != nil {
				return err
			}
			return query.Order("sort ASC").
				Limit(params.Size).
				Offset((params.Page - 1) * params.Size).
				Find(v).Error
		})
	}

	err := queryFunc(params.Language)
	if err != nil {
		return 0, nil, err
	}

	// fallback 默认语言
	if params.DefaultLanguage && total == 0 {
		err = queryFunc("")
		if err != nil {
			return 0, nil, err
		}
	}

	return total, list, nil
}

func InSet(field string, values []string) func(db *gorm.DB) *gorm.DB {
	return orm.CommaSeparatedContains(field, values)
}
