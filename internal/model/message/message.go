package message

// V4.3 user_message 站内信(决策 7.1)。
// 由 noticeDispatchLogic 写入(渲染好 title/body),用户中心拉取 + 标记已读。

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserMessage struct {
	Id        int64      `gorm:"primaryKey"`
	UserId    int64      `gorm:"column:user_id;index:idx_user_unread,priority:1;not null"`
	Category  string     `gorm:"column:category;type:varchar(32);not null;default:''"`
	Title     string     `gorm:"column:title;type:varchar(255);not null;default:''"`
	Body      string     `gorm:"column:body;type:mediumtext"`
	Link      string     `gorm:"column:link;type:varchar(512);not null;default:''"`
	ReadAt    *time.Time `gorm:"column:read_at;index:idx_user_unread,priority:2"`
	CreatedAt time.Time  `gorm:"<-:create"`
}

func (*UserMessage) TableName() string {
	return "user_message"
}

type Model interface {
	Insert(ctx context.Context, m *UserMessage) error
	List(ctx context.Context, userId int64, page, size int, unreadOnly bool) ([]*UserMessage, int64, error)
	UnreadCount(ctx context.Context, userId int64) (int64, error)
	MarkRead(ctx context.Context, userId, id int64) error
	MarkAllRead(ctx context.Context, userId int64) (int64, error)
}

type defaultModel struct {
	db *gorm.DB
}

func NewModel(db *gorm.DB) Model {
	return &defaultModel{db: db}
}

func (m *defaultModel) Insert(ctx context.Context, msg *UserMessage) error {
	return m.db.WithContext(ctx).Create(msg).Error
}

func (m *defaultModel) List(ctx context.Context, userId int64, page, size int, unreadOnly bool) ([]*UserMessage, int64, error) {
	if size <= 0 {
		size = 20
	}
	if page <= 0 {
		page = 1
	}
	conn := m.db.WithContext(ctx).Model(&UserMessage{}).Where("user_id = ?", userId)
	if unreadOnly {
		conn = conn.Where("read_at IS NULL")
	}
	var total int64
	if err := conn.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*UserMessage
	err := conn.Order("id DESC").Limit(size).Offset((page - 1) * size).Find(&list).Error
	return list, total, err
}

func (m *defaultModel) UnreadCount(ctx context.Context, userId int64) (int64, error) {
	var n int64
	err := m.db.WithContext(ctx).Model(&UserMessage{}).
		Where("user_id = ? AND read_at IS NULL", userId).Count(&n).Error
	return n, err
}

func (m *defaultModel) MarkRead(ctx context.Context, userId, id int64) error {
	now := time.Now()
	return m.db.WithContext(ctx).
		Model(&UserMessage{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL", id, userId).
		Update("read_at", now).Error
}

func (m *defaultModel) MarkAllRead(ctx context.Context, userId int64) (int64, error) {
	now := time.Now()
	res := m.db.WithContext(ctx).
		Model(&UserMessage{}).
		Where("user_id = ? AND read_at IS NULL", userId).
		Update("read_at", now)
	return res.RowsAffected, res.Error
}
