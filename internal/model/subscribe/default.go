package subscribe

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/pkg/cache"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customSubscribeModel)(nil)
var (
	cacheSubscribeIdPrefix = "cache:subscribe:id:"
)

type (
	Model interface {
		subscribeModel
		customSubscribeLogicModel
	}
	subscribeModel interface {
		Insert(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error
		FindOne(ctx context.Context, id int64) (*Subscribe, error)
		Update(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error
		Delete(ctx context.Context, id int64, tx ...*gorm.DB) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customSubscribeModel struct {
		*defaultSubscribeModel
	}
	defaultSubscribeModel struct {
		cache.CachedConn
		table string
	}
)

func newSubscribeModel(db *gorm.DB, c *redis.Client) *defaultSubscribeModel {
	return &defaultSubscribeModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`subscribe`",
	}
}

//nolint:unused
func (m *defaultSubscribeModel) batchGetCacheKeys(Subscribes ...*Subscribe) []string {
	var keys []string
	for _, subscribe := range Subscribes {
		keys = append(keys, m.getCacheKeys(subscribe)...)
	}
	return keys

}
func (m *defaultSubscribeModel) getCacheKeys(data *Subscribe) []string {
	if data == nil {
		return []string{}
	}
	var keys []string
	if data.Nodes != "" {
		var nodes []*node.Node
		ids := strings.Split(data.Nodes, ",")

		err := m.QueryNoCacheCtx(context.Background(), &nodes, func(conn *gorm.DB, v interface{}) error {
			return conn.Model(&node.Node{}).Where("id IN (?)", tool.StringSliceToInt64Slice(ids)).Find(&nodes).Error
		})
		if err == nil {
			for _, n := range nodes {
				keys = append(keys, fmt.Sprintf("%s%d", node.ServerUserListCacheKey, n.ServerId))
			}
		}
	}
	if data.NodeTags != "" {
		var nodes []*node.Node
		tags := tool.RemoveDuplicateElements(strings.Split(data.NodeTags, ",")...)
		err := m.QueryNoCacheCtx(context.Background(), &nodes, func(conn *gorm.DB, v interface{}) error {
			return conn.Model(&node.Node{}).Scopes(InSet("tags", tags)).Find(&nodes).Error
		})
		if err == nil {
			for _, n := range nodes {
				keys = append(keys, fmt.Sprintf("%s%d", node.ServerUserListCacheKey, n.ServerId))
			}
		}
	}

	return append(keys, fmt.Sprintf("%s%v", cacheSubscribeIdPrefix, data.Id))
}

func (m *defaultSubscribeModel) Insert(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultSubscribeModel) FindOne(ctx context.Context, id int64) (*Subscribe, error) {
	SubscribeIdKey := fmt.Sprintf("%s%v", cacheSubscribeIdPrefix, id)
	var resp Subscribe
	err := m.QueryCtx(ctx, &resp, SubscribeIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultSubscribeModel) Update(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error {
	old, err := m.FindOne(ctx, data.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		if len(tx) > 0 {
			db = tx[0]
		}
		return db.Save(data).Error
	}, m.getCacheKeys(old)...)
	return err
}

func (m *defaultSubscribeModel) Delete(ctx context.Context, id int64, tx ...*gorm.DB) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		if len(tx) > 0 {
			db = tx[0]
		}
		return db.Delete(&Subscribe{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultSubscribeModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
