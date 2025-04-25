package user

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (m *customUserModel) InsertSubscribeLog(ctx context.Context, log *SubscribeLog) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(log).Error
	})
}

func (m *customUserModel) FilterSubscribeLogList(ctx context.Context, page, size int, filter *SubscribeLogFilterParams) ([]*SubscribeLog, int64, error) {
	var list []*SubscribeLog
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		query := conn.Model(&SubscribeLog{})
		if filter != nil {
			if filter.UserId != 0 {
				query = query.Where("user_id = ?", filter.UserId)
			}
			if filter.UserSubscribeId != 0 {
				query = query.Where("user_subscribe_id = ?", filter.UserSubscribeId)
			}
			if filter.IP != "" {
				query = query.Where("ip LIKE ?", "%"+filter.IP+"%")
			}
			if filter.Token != "" {
				query = query.Where("token LIKE ?", "%"+filter.Token+"%")
			}
			if filter.UserAgent != "" {
				query = query.Where("user_agent LIKE ?", "%"+filter.UserAgent+"%")
			}
		}
		return query.Count(&total).Limit(size).Offset((page - 1) * size).Find(v).Error
	})

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, err
	}

	return list, total, nil
}

func (m *customUserModel) InsertLoginLog(ctx context.Context, log *LoginLog) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(log).Error
	})
}

func (m *customUserModel) FilterLoginLogList(ctx context.Context, page, size int, filter *LoginLogFilterParams) ([]*LoginLog, int64, error) {
	var list []*LoginLog
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		query := conn.Model(&LoginLog{})
		if filter != nil {
			if filter.UserId != 0 {
				query = query.Where("user_id = ?", filter.UserId)
			}
			if filter.IP != "" {
				query = query.Where("ip LIKE ?", "%"+filter.IP+"%")
			}
			if filter.UserAgent != "" {
				query = query.Where("user_agent LIKE ?", "%"+filter.UserAgent+"%")
			}
			if filter.Success != nil {
				query = query.Where("success = ?", *filter.Success)
			}
		}
		return query.Count(&total).Limit(size).Offset((page - 1) * size).Find(v).Error
	})

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, err
	}

	return list, total, nil
}
