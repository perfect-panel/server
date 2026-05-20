package user

import (
	"context"
	"time"

	"gorm.io/gorm"
)

func (m *customUserModel) FindTrafficExceededSubscribes(ctx context.Context) ([]*Subscribe, error) {
	var list []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).
			Where("upload + download >= traffic AND status IN ? AND traffic > 0", []int64{0, 1}).
			Find(&list).Error
	})
	return list, err
}

func (m *customUserModel) FindExpiredSubscribes(ctx context.Context, now time.Time) ([]*Subscribe, error) {
	var list []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).
			Where("status IN ? AND expire_time < ? AND expire_time != ? AND finished_at IS NULL", []int64{0, 1}, now, time.UnixMilli(0)).
			Find(&list).Error
	})
	return list, err
}

func (m *customUserModel) MarkSubscribesFinished(ctx context.Context, ids []int64, status uint8, finishedAt time.Time, tx ...*gorm.DB) error {
	if len(ids) == 0 {
		return nil
	}
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Model(&Subscribe{}).Where("id IN ?", ids).Updates(map[string]interface{}{
			"status":      status,
			"finished_at": finishedAt,
		}).Error
	})
}
