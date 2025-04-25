package ticket

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var cacheTicketDetailPrefix = "cache:ticket:detail:"

type Details struct {
	Id          int64     `gorm:"primaryKey"`
	Title       string    `gorm:"type:varchar(255);not null;default:'';comment:Title"`
	Description string    `gorm:"type:text;comment:Description"`
	UserId      int64     `gorm:"type:bigint;not null;default:0;comment:UserId"`
	Status      uint8     `gorm:"type:tinyint(1);not null;default:1;comment:Status"`
	Follows     []Follow  `gorm:"foreignKey:TicketId;references:Id"`
	CreatedAt   time.Time `gorm:"<-:create;comment:Create Time"`
	UpdatedAt   time.Time `gorm:"comment:Update Time"`
}
type customTicketLogicModel interface {
	QueryTicketDetail(ctx context.Context, id int64) (*Details, error)
	InsertTicketFollow(ctx context.Context, data *Follow) error
	QueryTicketList(ctx context.Context, page, size int, userId int64, status *uint8, search string) (int64, []*Ticket, error)
	UpdateTicketStatus(ctx context.Context, id, userId int64, status uint8) error
	QueryWaitReplyTotal(ctx context.Context) (int64, error)
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customTicketModel{
		defaultTicketModel: newTicketModel(conn, c),
	}
}

// QueryTicketDetail returns the ticket details.
func (m *customTicketModel) QueryTicketDetail(ctx context.Context, id int64) (*Details, error) {
	key := fmt.Sprintf("%s%v", cacheTicketDetailPrefix, id)
	var data *Details
	err := m.QueryCtx(ctx, &data, key, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Ticket{}).Where("id = ?", id).Preload("Follows").First(v).Error
	})
	return data, err
}

// InsertTicketFollow inserts a follow record.
func (m *customTicketModel) InsertTicketFollow(ctx context.Context, data *Follow) error {
	key := fmt.Sprintf("%s%v", cacheTicketDetailPrefix, data.TicketId)
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Model(&Follow{}).Create(data).Error
	}, key)
}

// QueryTicketList returns the ticket list.
func (m *customTicketModel) QueryTicketList(ctx context.Context, page, size int, userId int64, status *uint8, search string) (int64, []*Ticket, error) {
	var data []*Ticket
	var total int64
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		query := conn.Model(&Ticket{})
		if userId > 0 {
			query = query.Where("user_id = ?", userId)
		}
		if status != nil {
			query = query.Where("status = ?", status)
		} else {
			query = query.Where("status != ?", 4)
		}
		if search != "" {
			query = query.Where("title like ? or description like ?", "%"+search+"%", "%"+search+"%")
		}
		return query.Count(&total).Order("id desc").Limit(size).Offset((page - 1) * size).Find(v).Error
	})
	return total, data, err
}

// UpdateTicketStatus updates the ticket status.
func (m *customTicketModel) UpdateTicketStatus(ctx context.Context, id, userId int64, status uint8) error {
	key := fmt.Sprintf("%s%v", cacheTicketDetailPrefix, id)
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		conn = conn.Model(&Ticket{})
		if userId > 0 {
			conn = conn.Where("user_id = ?", userId)
		}
		return conn.Where("id = ?", id).Update("status", status).Error
	}, key)
}

// QueryWaitReplyTotal returns the total number of tickets that are waiting for a reply.
func (m *customTicketModel) QueryWaitReplyTotal(ctx context.Context) (int64, error) {
	var total int64
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Ticket{}).Where("status = ?", Pending).Count(&total).Error
	})
	return total, err
}
