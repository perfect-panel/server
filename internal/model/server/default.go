package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customServerModel)(nil)
var (
	cacheServerIdPrefix = "cache:server:id:"
)

type (
	Model interface {
		serverModel
		customServerLogicModel
	}
	serverModel interface {
		Insert(ctx context.Context, data *Server, tx ...*gorm.DB) error
		FindOne(ctx context.Context, id int64) (*Server, error)
		Update(ctx context.Context, data *Server, tx ...*gorm.DB) error
		Delete(ctx context.Context, id int64, tx ...*gorm.DB) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customServerModel struct {
		*defaultServerModel
	}
	defaultServerModel struct {
		cache.CachedConn
		table string
	}
)

func newServerModel(db *gorm.DB, c *redis.Client) *defaultServerModel {
	return &defaultServerModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`Server`",
	}
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customServerModel{
		defaultServerModel: newServerModel(conn, c),
	}
}

//nolint:unused
func (m *defaultServerModel) batchGetCacheKeys(Servers ...*Server) []string {
	var keys []string
	for _, server := range Servers {
		keys = append(keys, m.getCacheKeys(server)...)
	}
	return keys

}

func (m *defaultServerModel) getCacheKeys(data *Server) []string {
	if data == nil {
		return []string{}
	}
	detailsKey := fmt.Sprintf("%s%v", CacheServerDetailPrefix, data.Id)
	ServerIdKey := fmt.Sprintf("%s%v", cacheServerIdPrefix, data.Id)
	//configIdKey := fmt.Sprintf("%s%v", config.ServerConfigCacheKey, data.Id)
	//userIDKey := fmt.Sprintf("%s%d", config.ServerUserListCacheKey, data.Id)

	// query protocols to get config keys

	cacheKeys := []string{
		ServerIdKey,
		detailsKey,
		//configIdKey,
		//userIDKey,
	}
	return cacheKeys
}

func (m *defaultServerModel) Insert(ctx context.Context, data *Server, tx ...*gorm.DB) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultServerModel) FindOne(ctx context.Context, id int64) (*Server, error) {
	ServerIdKey := fmt.Sprintf("%s%v", cacheServerIdPrefix, id)
	var resp Server
	err := m.QueryCtx(ctx, &resp, ServerIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Server{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultServerModel) Update(ctx context.Context, data *Server, tx ...*gorm.DB) error {
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

func (m *defaultServerModel) Delete(ctx context.Context, id int64, tx ...*gorm.DB) error {
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
		return conn.Delete(&Server{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultServerModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
