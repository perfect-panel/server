package announcement

import (
	"context"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type customAnnouncementLogicModel interface {
	GetAnnouncementListByPage(ctx context.Context, page, size int, filter Filter) (int64, []*Announcement, error)
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customAnnouncementModel{
		defaultAnnouncementModel: newAnnouncementModel(conn, c),
	}
}

type Filter struct {
	Show   *bool
	Pinned *bool
	Popup  *bool
	Search string
}

// GetAnnouncementListByPage  get announcement list by page
func (m *customAnnouncementModel) GetAnnouncementListByPage(ctx context.Context, page, size int, filter Filter) (int64, []*Announcement, error) {
	var list []*Announcement
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.Model(&Announcement{})
		if filter.Show != nil {
			conn = conn.Where("`show` = ?", *filter.Show)
		}
		if filter.Pinned != nil {
			conn = conn.Where("`pinned` = ?", *filter.Pinned)
		}
		if filter.Popup != nil {
			conn = conn.Where("`popup` = ?", *filter.Popup)
		}
		if filter.Search != "" {
			conn = conn.Where("`title` LIKE ? OR `content` LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
		}
		return conn.Count(&total).Offset((page - 1) * size).Limit(size).Find(&list).Error
	})
	return total, list, err
}
