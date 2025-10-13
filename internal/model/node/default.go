package node

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customServerModel)(nil)

//goland:noinspection GoNameStartsWithPackageName
type (
	Model interface {
		serverModel
		NodeModel
		customCacheLogicModel
		customServerLogicModel
	}
	serverModel interface {
		InsertServer(ctx context.Context, data *Server, tx ...*gorm.DB) error
		FindOneServer(ctx context.Context, id int64) (*Server, error)
		UpdateServer(ctx context.Context, data *Server, tx ...*gorm.DB) error
		DeleteServer(ctx context.Context, id int64, tx ...*gorm.DB) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
		QueryServerList(ctx context.Context, ids []int64) (servers []*Server, err error)
	}

	NodeModel interface {
		InsertNode(ctx context.Context, data *Node, tx ...*gorm.DB) error
		FindOneNode(ctx context.Context, id int64) (*Node, error)
		UpdateNode(ctx context.Context, data *Node, tx ...*gorm.DB) error
		DeleteNode(ctx context.Context, id int64, tx ...*gorm.DB) error
	}

	customServerModel struct {
		*defaultServerModel
	}
	defaultServerModel struct {
		*gorm.DB
		Cache *redis.Client
	}
)

func newServerModel(db *gorm.DB, cache *redis.Client) *defaultServerModel {
	return &defaultServerModel{
		DB:    db,
		Cache: cache,
	}
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, cache *redis.Client) Model {
	return &customServerModel{
		defaultServerModel: newServerModel(conn, cache),
	}
}

func (m *defaultServerModel) InsertServer(ctx context.Context, data *Server, tx ...*gorm.DB) error {
	db := m.DB
	if len(tx) > 0 {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(data).Error
}

func (m *defaultServerModel) FindOneServer(ctx context.Context, id int64) (*Server, error) {
	var server Server
	err := m.WithContext(ctx).Model(&Server{}).Where("id = ?", id).First(&server).Error
	return &server, err
}

func (m *defaultServerModel) UpdateServer(ctx context.Context, data *Server, tx ...*gorm.DB) error {
	_, err := m.FindOneServer(ctx, data.Id)
	if err != nil {
		return err
	}

	db := m.DB
	if len(tx) > 0 {
		db = tx[0]
	}
	return db.WithContext(ctx).Where("`id` = ?", data.Id).Save(data).Error

}

func (m *defaultServerModel) DeleteServer(ctx context.Context, id int64, tx ...*gorm.DB) error {
	db := m.DB
	if len(tx) > 0 {
		db = tx[0]
	}
	return db.WithContext(ctx).Where("`id` = ?", id).Delete(&Server{}).Error
}

func (m *defaultServerModel) InsertNode(ctx context.Context, data *Node, tx ...*gorm.DB) error {
	db := m.DB
	if len(tx) > 0 {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(data).Error
}

func (m *defaultServerModel) FindOneNode(ctx context.Context, id int64) (*Node, error) {
	var node Node
	err := m.WithContext(ctx).Model(&Node{}).Where("id = ?", id).First(&node).Error
	return &node, err
}

func (m *defaultServerModel) UpdateNode(ctx context.Context, data *Node, tx ...*gorm.DB) error {
	_, err := m.FindOneNode(ctx, data.Id)
	if err != nil {
		return err
	}

	db := m.DB
	if len(tx) > 0 {
		db = tx[0]
	}
	return db.WithContext(ctx).Where("`id` = ?", data.Id).Save(data).Error
}

func (m *defaultServerModel) DeleteNode(ctx context.Context, id int64, tx ...*gorm.DB) error {
	db := m.DB
	if len(tx) > 0 {
		db = tx[0]
	}
	return db.WithContext(ctx).Where("`id` = ?", id).Delete(&Node{}).Error
}

func (m *defaultServerModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.WithContext(ctx).Transaction(fn)
}
