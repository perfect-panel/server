package user

import (
	"context"

	"gorm.io/gorm"
)

func (m *customUserModel) CountAffiliates(ctx context.Context, refererId int64) (int64, error) {
	var total int64
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Where("referer_id = ?", refererId).Count(&total).Error
	})
	return total, err
}

func (m *customUserModel) QueryAffiliateList(ctx context.Context, refererId int64, page, size int) ([]*User, int64, error) {
	var list []*User
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).
			Where("referer_id = ?", refererId).
			Count(&total).
			Order("id desc").
			Limit(size).
			Offset((page - 1) * size).
			Preload("AuthMethods").
			Find(&list).Error
	})
	return list, total, err
}
