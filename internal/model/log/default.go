package log

import (
	"context"

	"gorm.io/gorm"
)

var _ Model = (*customLogModel)(nil)

type (
	Model interface {
		messageLogModel
	}
	messageLogModel interface {
		InsertMessageLog(ctx context.Context, data *MessageLog) error
		FindOneMessageLog(ctx context.Context, id int64) (*MessageLog, error)
		UpdateMessageLog(ctx context.Context, data *MessageLog) error
		DeleteMessageLog(ctx context.Context, id int64) error
		FindMessageLogList(ctx context.Context, page, size int, filter MessageLogFilterParams) (int64, []*MessageLog, error)
	}

	customLogModel struct {
		*defaultLogModel
	}
	defaultLogModel struct {
		Connection *gorm.DB
	}
)

func newLogModel(db *gorm.DB) *defaultLogModel {
	return &defaultLogModel{
		Connection: db,
	}
}

func (m *defaultLogModel) InsertMessageLog(ctx context.Context, data *MessageLog) error {
	return m.Connection.WithContext(ctx).Create(&data).Error
}

func (m *defaultLogModel) FindOneMessageLog(ctx context.Context, id int64) (*MessageLog, error) {
	var resp MessageLog
	err := m.Connection.WithContext(ctx).Model(&MessageLog{}).Where("`id` = ?", id).First(&resp).Error
	return &resp, err
}

func (m *defaultLogModel) UpdateMessageLog(ctx context.Context, data *MessageLog) error {
	return m.Connection.WithContext(ctx).Model(&MessageLog{}).Where("id = ?", data.Id).Updates(data).Error
}

func (m *defaultLogModel) DeleteMessageLog(ctx context.Context, id int64) error {
	return m.Connection.WithContext(ctx).Model(&MessageLog{}).Where("id = ?", id).Delete(&MessageLog{}).Error
}

func (m *defaultLogModel) FindMessageLogList(ctx context.Context, page, size int, filter MessageLogFilterParams) (int64, []*MessageLog, error) {
	var list []*MessageLog
	var total int64
	conn := m.Connection.WithContext(ctx).Model(&MessageLog{})
	if filter.Type != "" {
		conn = conn.Where("`type` = ?", filter.Type)
	}
	if filter.Platform != "" {
		conn = conn.Where("`platform` = ?", filter.Platform)
	}
	if filter.To != "" {
		conn = conn.Where("`to` LIKE ?", "%"+filter.To+"%")
	}
	if filter.Subject != "" {
		conn = conn.Where("`subject` LIKE ?", "%"+filter.Subject+"%")
	}
	if filter.Content != "" {
		conn = conn.Where("`content` = ?", "%"+filter.Content+"%")
	}
	if filter.Status > 0 {
		conn = conn.Where("`status` = ?", filter.Status)
	}
	err := conn.Count(&total).Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}
