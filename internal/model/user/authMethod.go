package user

import (
	"context"

	"github.com/perfect-panel/server/pkg/logger"
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
	u, err := m.FindOne(ctx, data.UserId)
	if err != nil {
		return err
	}

	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		if err = conn.Model(&AuthMethods{}).Create(data).Error; err != nil {
			return err
		}
		return m.ClearUserCache(ctx, u)
	})
}

func (m *defaultUserModel) UpdateUserAuthMethods(ctx context.Context, data *AuthMethods, tx ...*gorm.DB) error {
	u, err := m.FindOne(ctx, data.UserId)
	if err != nil {
		return err
	}

	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		err = conn.Model(&AuthMethods{}).Where("user_id = ? AND auth_type = ?", data.UserId, data.AuthType).Save(data).Error
		if err != nil {
			return err
		}
		return m.ClearUserCache(ctx, u)
	})
}

func (m *defaultUserModel) DeleteUserAuthMethods(ctx context.Context, userId int64, platform string, tx ...*gorm.DB) error {
	u, err := m.FindOne(ctx, userId)
	if err != nil {
		return err
	}
	defer func() {
		if err = m.ClearUserCache(context.Background(), u); err != nil {
			logger.Errorf("[UserModel] clear user cache failed: %v", err.Error())
		}
	}()
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
