package log

import (
	"context"

	"gorm.io/gorm"
)

func NewModel(db *gorm.DB) Model {
	return &customSystemLogModel{
		defaultLogModel: newSystemLogModel(db),
	}
}

type FilterParams struct {
	Page     int
	Size     int
	Type     uint8
	Data     string
	Search   string
	ObjectID int64
}

type customSystemLogLogicModel interface {
	FilterSystemLog(ctx context.Context, filter *FilterParams) ([]*SystemLog, int64, error)
}

func (m *customSystemLogModel) FilterSystemLog(ctx context.Context, filter *FilterParams) ([]*SystemLog, int64, error) {
	tx := m.WithContext(ctx).Model(&SystemLog{}).Order("id DESC")
	if filter == nil {
		filter = &FilterParams{
			Page: 1,
			Size: 10,
		}
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Size < 1 {
		filter.Size = 10
	}

	if filter.Type != 0 {
		tx = tx.Where("`type` = ?", filter.Type)
	}

	if filter.Data != "" {
		tx = tx.Where("`date` = ?", filter.Data)
	}

	if filter.ObjectID != 0 {
		tx = tx.Where("`object_id` = ?", filter.ObjectID)
	}
	if filter.Search != "" {
		tx = tx.Where("`content` LIKE ?", "%"+filter.Search+"%")
	}

	var total int64
	var logs []*SystemLog
	err := tx.Count(&total).Limit(filter.Size).Offset((filter.Page - 1) * filter.Size).Find(&logs).Error
	return logs, total, err
}
