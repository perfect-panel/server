package client

import (
	"context"

	"gorm.io/gorm"
)

type (
	Model interface {
		subscribeApplicationModel
	}
	subscribeApplicationModel interface {
		Insert(ctx context.Context, data *SubscribeApplication) error
		FindOne(ctx context.Context, id int64) (*SubscribeApplication, error)
		Update(ctx context.Context, data *SubscribeApplication) error
		Delete(ctx context.Context, id int64) error
		List(ctx context.Context) ([]*SubscribeApplication, error)
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}
	DefaultSubscribeApplicationModel struct {
		*gorm.DB
	}
)

func NewSubscribeApplicationModel(db *gorm.DB) Model {
	return &DefaultSubscribeApplicationModel{
		DB: db,
	}
}

func (m *DefaultSubscribeApplicationModel) Insert(ctx context.Context, data *SubscribeApplication) error {
	if err := m.WithContext(ctx).Model(&SubscribeApplication{}).Create(data).Error; err != nil {
		return err
	}
	return nil
}

func (m *DefaultSubscribeApplicationModel) FindOne(ctx context.Context, id int64) (*SubscribeApplication, error) {
	var resp SubscribeApplication
	if err := m.WithContext(ctx).Model(&SubscribeApplication{}).Where("id = ?", id).First(&resp).Error; err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *DefaultSubscribeApplicationModel) Update(ctx context.Context, data *SubscribeApplication) error {
	if _, err := m.FindOne(ctx, data.Id); err != nil {
		return err
	}
	if err := m.WithContext(ctx).Model(&SubscribeApplication{}).Where("`id` = ?", data.Id).Save(data).Error; err != nil {
		return err
	}
	return nil
}

func (m *DefaultSubscribeApplicationModel) Delete(ctx context.Context, id int64) error {
	if err := m.WithContext(ctx).Model(&SubscribeApplication{}).Where("`id` = ?", id).Delete(&SubscribeApplication{}).Error; err != nil {
		return err
	}
	return nil
}

func (m *DefaultSubscribeApplicationModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	tx := m.WithContext(ctx).Begin()
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return rbErr
		}
		return err
	}
	return tx.Commit().Error
}

func (m *DefaultSubscribeApplicationModel) List(ctx context.Context) ([]*SubscribeApplication, error) {
	var resp []*SubscribeApplication
	if err := m.WithContext(ctx).Find(&resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}
