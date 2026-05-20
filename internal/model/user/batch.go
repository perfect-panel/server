package user

import (
	"context"

	"gorm.io/gorm"
)

func (m *customUserModel) FindUsersByIds(ctx context.Context, ids []int64) ([]*User, error) {
	var users []*User
	if len(ids) == 0 {
		return users, nil
	}
	err := m.QueryNoCacheCtx(ctx, &users, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Where("id IN ?", ids).Find(&users).Error
	})
	return users, err
}

func (m *customUserModel) FindSubscribesByIds(ctx context.Context, ids []int64) ([]*Subscribe, error) {
	var subscribes []*Subscribe
	if len(ids) == 0 {
		return subscribes, nil
	}
	err := m.QueryNoCacheCtx(ctx, &subscribes, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("id IN ?", ids).Find(&subscribes).Error
	})
	return subscribes, err
}
