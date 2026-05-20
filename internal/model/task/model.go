package task

import (
	"context"

	"gorm.io/gorm"
)

var _ Model = (*defaultTaskModel)(nil)

type Model interface {
	Insert(ctx context.Context, data *Task) error
	FindOne(ctx context.Context, id int64) (*Task, error)
	FindOneByType(ctx context.Context, id int64, typ Type) (*Task, error)
	QueryTaskList(ctx context.Context, filter *Filter) (int64, []*Task, error)
	UpdateStatus(ctx context.Context, id int64, status int8) error
}

type Filter struct {
	Type   Type
	Page   int
	Size   int
	Status *uint8
	Scope  *int8
}

type defaultTaskModel struct {
	db *gorm.DB
}

func NewModel(db *gorm.DB) Model {
	return &defaultTaskModel{
		db: db,
	}
}

func (m *defaultTaskModel) Insert(ctx context.Context, data *Task) error {
	return m.db.WithContext(ctx).Create(data).Error
}

func (m *defaultTaskModel) FindOne(ctx context.Context, id int64) (*Task, error) {
	var data Task
	err := m.db.WithContext(ctx).Model(&Task{}).Where("id = ?", id).First(&data).Error
	return &data, err
}

func (m *defaultTaskModel) FindOneByType(ctx context.Context, id int64, typ Type) (*Task, error) {
	var data Task
	err := m.db.WithContext(ctx).Model(&Task{}).Where("id = ? AND type = ?", id, typ).First(&data).Error
	return &data, err
}

func (m *defaultTaskModel) QueryTaskList(ctx context.Context, filter *Filter) (int64, []*Task, error) {
	var total int64
	var data []*Task
	if filter == nil {
		filter = &Filter{
			Type: Undefined,
			Page: 1,
			Size: 10,
		}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Size <= 0 {
		filter.Size = 10
	}

	query := m.db.WithContext(ctx).Model(&Task{})
	if filter.Type != Undefined {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Scope != nil {
		var all []*Task
		if err := query.Order("created_at DESC").Find(&all).Error; err != nil {
			return 0, nil, err
		}

		// Scope is stored as JSON text; filter here to keep the query dialect-neutral.
		filtered := make([]*Task, 0, len(all))
		for _, item := range all {
			var scope EmailScope
			if err := scope.Unmarshal([]byte(item.Scope)); err != nil {
				continue
			}
			if scope.Type == *filter.Scope {
				filtered = append(filtered, item)
			}
		}

		total = int64(len(filtered))
		start := (filter.Page - 1) * filter.Size
		if start >= len(filtered) {
			return total, []*Task{}, nil
		}
		end := start + filter.Size
		if end > len(filtered) {
			end = len(filtered)
		}
		return total, filtered[start:end], nil
	}

	err := query.Count(&total).
		Offset((filter.Page - 1) * filter.Size).
		Limit(filter.Size).
		Order("created_at DESC").
		Find(&data).Error
	return total, data, err
}

func (m *defaultTaskModel) UpdateStatus(ctx context.Context, id int64, status int8) error {
	return m.db.WithContext(ctx).Model(&Task{}).Where("id = ?", id).Update("status", status).Error
}
