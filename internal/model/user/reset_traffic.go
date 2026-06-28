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
		return monthlyResetSubscribeQuery(conn, subscribeIds, now).Find(&ids).Error
	})
	return ids, err
}

func (m *customUserModel) QueryFirstResetSubscribeIds(ctx context.Context, subscribeIds []int64, now time.Time) ([]int64, error) {
	var ids []int64
	if len(subscribeIds) == 0 {
		return ids, nil
	}
	err := m.QueryNoCacheCtx(ctx, &ids, func(conn *gorm.DB, v interface{}) error {
		return resettableSubscribeQuery(conn, subscribeIds, now).Find(&ids).Error
	})
	return ids, err
}

func (m *customUserModel) QueryYearlyResetSubscribeIds(ctx context.Context, subscribeIds []int64, now time.Time) ([]int64, error) {
	var ids []int64
	if len(subscribeIds) == 0 {
		return ids, nil
	}
	err := m.QueryNoCacheCtx(ctx, &ids, func(conn *gorm.DB, v interface{}) error {
		return yearlyResetSubscribeQuery(conn, subscribeIds, now).Find(&ids).Error
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

func extractColumnDatePart(db *gorm.DB, column, part string) string {
	if db.Dialector.Name() == "postgres" {
		return fmt.Sprintf("EXTRACT(%s FROM %s)", part, column)
	}
	switch part {
	case "month":
		return fmt.Sprintf("MONTH(%s)", column)
	default:
		return fmt.Sprintf("DAY(%s)", column)
	}
}

func monthlyResetSubscribeQuery(conn *gorm.DB, subscribeIds []int64, now time.Time) *gorm.DB {
	query := resettableSubscribeQuery(conn, subscribeIds, now)
	condition, args := monthlyResetDateCondition(conn, now)
	return query.Where(condition, args...)
}

func yearlyResetSubscribeQuery(conn *gorm.DB, subscribeIds []int64, now time.Time) *gorm.DB {
	query := resettableSubscribeQuery(conn, subscribeIds, now)
	condition, args := yearlyResetDateCondition(conn, now)
	return query.Where(condition, args...)
}

func resettableSubscribeQuery(conn *gorm.DB, subscribeIds []int64, now time.Time) *gorm.DB {
	return conn.Model(&Subscribe{}).Select("id").
		Where("subscribe_id IN ?", subscribeIds).
		Where("status IN ?", []int64{1, 2}).
		Where("start_time <= ?", now).
		Where("(expire_time IS NULL OR expire_time = ? OR expire_time > ?)", time.UnixMilli(0), now)
}

func monthlyResetDateCondition(db *gorm.DB, now time.Time) (string, []interface{}) {
	dayExpr := extractColumnDatePart(db, "start_time", "day")
	if isLastDayOfMonth(now) {
		return dayExpr + " >= ?", []interface{}{now.Day()}
	}
	return dayExpr + " = ?", []interface{}{now.Day()}
}

func yearlyResetDateCondition(db *gorm.DB, now time.Time) (string, []interface{}) {
	monthExpr := extractColumnDatePart(db, "start_time", "month")
	dayExpr := extractColumnDatePart(db, "start_time", "day")
	if now.Month() == time.February && now.Day() == 28 && !isLeapYear(now.Year()) {
		return fmt.Sprintf("%s = ? AND %s IN ?", monthExpr, dayExpr), []interface{}{int(time.February), []int{28, 29}}
	}
	return fmt.Sprintf("%s = ? AND %s = ?", monthExpr, dayExpr), []interface{}{int(now.Month()), now.Day()}
}

func isLastDayOfMonth(t time.Time) bool {
	return t.AddDate(0, 0, 1).Month() != t.Month()
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
