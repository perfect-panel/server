package document

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customDocumentModel)(nil)
var (
	cacheDocumentIdPrefix = "cache:document:id:"
)

type (
	Model interface {
		documentModel
		customDocumentLogicModel
	}
	documentModel interface {
		Insert(ctx context.Context, data *Document) error
		FindOne(ctx context.Context, id int64) (*Document, error)
		Update(ctx context.Context, data *Document) error
		Delete(ctx context.Context, id int64) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customDocumentModel struct {
		*defaultDocumentModel
	}
	defaultDocumentModel struct {
		cache.CachedConn
		table string
	}
)

func newDocumentModel(db *gorm.DB, c *redis.Client) *defaultDocumentModel {
	return &defaultDocumentModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`document`",
	}
}

//nolint:unused
func (m *defaultDocumentModel) batchGetCacheKeys(Documents ...*Document) []string {
	var keys []string
	for _, document := range Documents {
		keys = append(keys, m.getCacheKeys(document)...)
	}
	return keys

}
func (m *defaultDocumentModel) getCacheKeys(data *Document) []string {
	if data == nil {
		return []string{}
	}
	documentIdKey := fmt.Sprintf("%s%v", cacheDocumentIdPrefix, data.Id)
	cacheKeys := []string{
		documentIdKey,
	}
	return cacheKeys
}

func (m *defaultDocumentModel) Insert(ctx context.Context, data *Document) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultDocumentModel) FindOne(ctx context.Context, id int64) (*Document, error) {
	DocumentIdKey := fmt.Sprintf("%s%v", cacheDocumentIdPrefix, id)
	var resp Document
	err := m.QueryCtx(ctx, &resp, DocumentIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Document{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultDocumentModel) Update(ctx context.Context, data *Document) error {
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

func (m *defaultDocumentModel) Delete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Delete(&Document{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultDocumentModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
