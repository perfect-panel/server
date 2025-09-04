package subscribe

import (
	"context"

	"github.com/perfect-panel/server/pkg/tool"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
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
			s := "%" + params.Search + "%"
			query = query.Where("`name` LIKE ? OR `description` LIKE ?", s, s)
		}
		if params.Show {
			query = query.Where("`show` = true")
		}
		if params.Sell {
			query = query.Where("`sell` = true")
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
	return func(db *gorm.DB) *gorm.DB {
		if len(values) == 0 {
			return db
		}

		query := db.Where("1=0")
		for _, v := range values {
			query = query.Or("FIND_IN_SET(?, "+field+")", v)
		}
		return query
	}
}
