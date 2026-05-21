package auth

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type customAuthLogicModel interface {
	GetAuthListByPage(ctx context.Context) ([]*Auth, error)
	FindOneByMethod(ctx context.Context, platform string) (*Auth, error)
	FindAll(ctx context.Context) ([]*Auth, error)
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customAuthModel{
		defaultAuthModel: newAuthModel(conn, c),
	}
}

type Filter struct {
	Show   *bool
	Pinned *bool
	Popup  *bool
	Search string
}

// GetAuthListByPage  get auth list by page
func (m *customAuthModel) GetAuthListByPage(ctx context.Context) ([]*Auth, error) {
	var list []*Auth
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.Model(&Auth{})
		return conn.Find(v).Error
	})
	return list, err
}

// FindOneByMethod  find one by method
func (m *customAuthModel) FindOneByMethod(ctx context.Context, method string) (*Auth, error) {
	var data Auth
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Auth{}).Where("method = ?", method).First(v).Error
	})

	return &data, err
}

// FindAll find all
func (m *customAuthModel) FindAll(ctx context.Context) ([]*Auth, error) {
	var list []*Auth
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.Model(&Auth{})
		return conn.Find(v).Error
	})
	return list, err
}
