package log

import (
	"context"

	"gorm.io/gorm"
)

var _ Model = (*customSystemLogModel)(nil)

type (
	Model interface {
		systemLogModel
		customSystemLogLogicModel
	}
	systemLogModel interface {
		Insert(ctx context.Context, data *SystemLog) error
		FindOne(ctx context.Context, id int64) (*SystemLog, error)
		Update(ctx context.Context, data *SystemLog) error
		Delete(ctx context.Context, id int64) error
	}
	customSystemLogModel struct {
		*defaultLogModel
	}
	defaultLogModel struct {
		*gorm.DB
	}
)

func newSystemLogModel(db *gorm.DB) *defaultLogModel {
	return &defaultLogModel{
		DB: db,
	}
}

func (m *defaultLogModel) Insert(ctx context.Context, data *SystemLog) error {
	return m.WithContext(ctx).Create(data).Error
}

func (m *defaultLogModel) FindOne(ctx context.Context, id int64) (*SystemLog, error) {
	var log SystemLog
	err := m.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (m *defaultLogModel) Update(ctx context.Context, data *SystemLog) error {
	return m.WithContext(ctx).Where("`id` = ?", data.Id).Save(data).Error
}

func (m *defaultLogModel) Delete(ctx context.Context, id int64) error {
	return m.WithContext(ctx).Where("`id` = ?", id).Delete(&SystemLog{}).Error
}
