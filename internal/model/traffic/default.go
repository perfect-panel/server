package traffic

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var _ Model = (*customTrafficModel)(nil)

type (
	Model interface {
		trafficModel
		customTrafficLogicModel
	}
	trafficModel interface {
		Insert(ctx context.Context, data *TrafficLog) error
		FindOne(ctx context.Context, id int64) (*TrafficLog, error)
		Update(ctx context.Context, data *TrafficLog) error
		Delete(ctx context.Context, id int64) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customTrafficModel struct {
		*defaultTrafficModel
	}
	defaultTrafficModel struct {
		Conn  *gorm.DB
		table string
	}
)

func newTrafficModel(db *gorm.DB) *defaultTrafficModel {
	return &defaultTrafficModel{
		Conn:  db,
		table: "`traffic`",
	}
}

func (m *defaultTrafficModel) Insert(ctx context.Context, data *TrafficLog) error {
	return m.Conn.WithContext(ctx).Create(&data).Error
}

func (m *defaultTrafficModel) FindOne(ctx context.Context, id int64) (*TrafficLog, error) {
	var data TrafficLog
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).Where("`id` = ?", id).First(&data).Error
	return &data, err
}

func (m *defaultTrafficModel) Update(ctx context.Context, data *TrafficLog) error {
	return m.Conn.WithContext(ctx).Save(data).Error
}

func (m *defaultTrafficModel) Delete(ctx context.Context, id int64) error {
	_, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return m.Conn.WithContext(ctx).Delete(&TrafficLog{}, id).Error
}

func (m *defaultTrafficModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.Conn.WithContext(ctx).Transaction(fn)
}
