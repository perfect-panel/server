package user

import (
	"context"

	"gorm.io/gorm"
)

func (m *defaultUserModel) FindUserAuthMethods(ctx context.Context, userId int64) ([]*AuthMethods, error) {
	var data []*AuthMethods
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&AuthMethods{}).Where("user_id = ?", userId).Find(&data).Error
	})
	return data, err
}

func (m *defaultUserModel) FindUserAuthMethodByOpenID(ctx context.Context, method, openID string) (*AuthMethods, error) {
	var data AuthMethods
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&AuthMethods{}).Where("auth_type = ? AND auth_identifier = ?", method, openID).First(&data).Error
	})
	return &data, err
}

func (m *defaultUserModel) FindUserAuthMethodByPlatform(ctx context.Context, userId int64, platform string) (*AuthMethods, error) {
	var data AuthMethods
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&AuthMethods{}).Where("user_id = ? AND auth_type = ?", userId, platform).First(&data).Error
	})
	return &data, err
}

func (m *defaultUserModel) InsertUserAuthMethods(ctx context.Context, data *AuthMethods, tx ...*gorm.DB) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Model(&AuthMethods{}).Create(data).Error
	})
}

func (m *defaultUserModel) UpdateUserAuthMethods(ctx context.Context, data *AuthMethods, tx ...*gorm.DB) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Model(&AuthMethods{}).Where("user_id = ? AND auth_type = ?", data.UserId, data.AuthType).Save(data).Error
	})
}

func (m *defaultUserModel) DeleteUserAuthMethods(ctx context.Context, userId int64, platform string, tx ...*gorm.DB) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Model(&AuthMethods{}).Where("user_id = ? AND auth_type = ?", userId, platform).Delete(&AuthMethods{}).Error
	})
}

func (m *defaultUserModel) FindUserAuthMethodByUserId(ctx context.Context, method string, userId int64) (*AuthMethods, error) {
	var data AuthMethods
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&AuthMethods{}).Where("auth_type = ? AND user_id = ?", method, userId).First(&data).Error
	})
	return &data, err
}
