package user

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (m *customUserModel) QueryMonthlyResetSubscribeIds(ctx context.Context, subscribeIds []int64, now time.Time) ([]int64, error) {
	var ids []int64
	if len(subscribeIds) == 0 {
		return ids, nil
	}
	err := m.QueryNoCacheCtx(ctx, &ids, func(conn *gorm.DB, v interface{}) error {
		query := conn.Model(&Subscribe{}).Select("id").
			Where("subscribe_id IN ?", subscribeIds).
			Where("status IN ?", []int64{1, 2}).
			Where(expireDateAtLeast(conn, "month"))

		if now.AddDate(0, 0, 1).Month() != now.Month() {
			query = query.Where(extractDatePart(conn, "day")+" >= ?", now.Day())
		} else {
			query = query.Where(extractDatePart(conn, "day")+" = ?", now.Day())
		}

		return query.Find(&ids).Error
	})
	return ids, err
}

func (m *customUserModel) QueryFirstResetSubscribeIds(ctx context.Context, subscribeIds []int64) ([]int64, error) {
	var ids []int64
	if len(subscribeIds) == 0 {
		return ids, nil
	}
	err := m.QueryNoCacheCtx(ctx, &ids, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Select("id").
			Where("subscribe_id IN ?", subscribeIds).
			Where("status IN ?", []int64{1, 2}).
			Find(&ids).Error
	})
	return ids, err
}

func (m *customUserModel) QueryYearlyResetSubscribeIds(ctx context.Context, subscribeIds []int64, now time.Time) ([]int64, error) {
	var ids []int64
	if len(subscribeIds) == 0 {
		return ids, nil
	}
	err := m.QueryNoCacheCtx(ctx, &ids, func(conn *gorm.DB, v interface{}) error {
		query := conn.Model(&Subscribe{}).Select("id").
			Where("subscribe_id IN ?", subscribeIds).
			Where(extractDatePart(conn, "month")+" = ?", int(now.Month())).
			Where("status IN ?", []int64{1, 2}).
			Where(expireDateAtLeast(conn, "year"))

		if now.Month() == 2 && now.Day() == 28 {
			query = query.Where(extractDatePart(conn, "day") + " IN (28, 29)")
		} else {
			query = query.Where(extractDatePart(conn, "day")+" = ?", now.Day())
		}

		return query.Find(&ids).Error
	})
	return ids, err
}

func (m *customUserModel) ResetSubscribeTrafficByIds(ctx context.Context, ids []int64, tx ...*gorm.DB) error {
	if len(ids) == 0 {
		return nil
	}
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Model(&Subscribe{}).Where("id IN ?", ids).
			Updates(map[string]interface{}{
				"upload":      0,
				"download":    0,
				"status":      1,
				"finished_at": nil,
			}).Error
	})
}

func extractDatePart(db *gorm.DB, part string) string {
	if db.Dialector.Name() == "postgres" {
		return fmt.Sprintf("EXTRACT(%s FROM expire_time)", part)
	}
	switch part {
	case "month":
		return "MONTH(expire_time)"
	default:
		return "DAY(expire_time)"
	}
}

func expireDateAtLeast(db *gorm.DB, unit string) string {
	if db.Dialector.Name() == "postgres" {
		return fmt.Sprintf("DATE(expire_time) >= CURRENT_DATE + INTERVAL '1 %s'", unit)
	}
	switch unit {
	case "year":
		return "TIMESTAMPDIFF(YEAR, CURDATE(), DATE(expire_time)) >= 1"
	default:
		return "TIMESTAMPDIFF(MONTH, CURDATE(), DATE(expire_time)) >= 1"
	}
}
