package traffic

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type customTrafficLogicModel interface {
	QueryServerTrafficByDay(ctx context.Context, serverId int64, date time.Time) (*TotalTraffic, error)
	QueryTrafficByDay(ctx context.Context, date time.Time) (*TotalTraffic, error)
	QueryTrafficByMonthly(ctx context.Context, date time.Time) (*TotalTraffic, error)
	TopServersTrafficByDay(ctx context.Context, date time.Time, limit int) ([]ServerTrafficRanking, error)
	TopServersTrafficByMonthly(ctx context.Context, date time.Time, limit int) ([]ServerTrafficRanking, error)
	TopUsersTrafficByDay(ctx context.Context, date time.Time, limit int) ([]UserTrafficRanking, error)
	TopUsersTrafficByMonthly(ctx context.Context, date time.Time, limit int) ([]UserTrafficRanking, error)
	QueryTrafficLogPageList(ctx context.Context, userId, subscribeId int64, page, size int) ([]*TrafficLog, int64, error)
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB) Model {
	return &customTrafficModel{
		defaultTrafficModel: newTrafficModel(conn),
	}
}

func (m *customTrafficModel) QueryServerTrafficByDay(ctx context.Context, serverId int64, date time.Time) (*TotalTraffic, error) {
	var data TotalTraffic
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select("sum(download) as download, sum(upload) as upload").
		Where("server_id = ? AND timestamp BETWEEN ? AND ?", serverId, start, end).
		Scan(&data).Error
	return &data, err
}

func (m *customTrafficModel) QueryTrafficByDay(ctx context.Context, date time.Time) (*TotalTraffic, error) {
	var data TotalTraffic
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select("sum(download) as download, sum(upload) as upload").
		Where("timestamp BETWEEN ? AND ?", start, end).
		Scan(&data).Error
	return &data, err
}

func (m *customTrafficModel) QueryTrafficByMonthly(ctx context.Context, date time.Time) (*TotalTraffic, error) {
	var data TotalTraffic
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select("sum(download) as download, sum(upload) as upload").
		Where("timestamp BETWEEN ? AND ?", start, end).
		Scan(&data).Error
	return &data, err
}

func (m *customTrafficModel) TopServersTrafficByDay(ctx context.Context, date time.Time, limit int) ([]ServerTrafficRanking, error) {
	var summaries []ServerTrafficRanking
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)
	err := m.Conn.Debug().WithContext(ctx).Model(&TrafficLog{}).
		Select("server_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
		Where("timestamp BETWEEN ? AND ?", start, end).
		Group("server_id").
		Order("total DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

func (m *customTrafficModel) TopServersTrafficByMonthly(ctx context.Context, date time.Time, limit int) ([]ServerTrafficRanking, error) {
	var summaries []ServerTrafficRanking
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select("server_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
		Where("timestamp BETWEEN ? AND ?", start, end).
		Group("server_id").
		Order("total DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

func (m *customTrafficModel) TopUsersTrafficByDay(ctx context.Context, date time.Time, limit int) ([]UserTrafficRanking, error) {
	var summaries []UserTrafficRanking
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select("user_id, subscribe_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
		Where("timestamp BETWEEN ? AND ?", start, end).
		Group("user_id, subscribe_id"). // 修改这里，添加 subscribe_id 到 GROUP BY 子句
		Order("total DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

func (m *customTrafficModel) TopUsersTrafficByMonthly(ctx context.Context, date time.Time, limit int) ([]UserTrafficRanking, error) {
	var summaries []UserTrafficRanking
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select("user_id, subscribe_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload"). // 添加 subscribe_id 到 SELECT 列表
		Where("timestamp BETWEEN ? AND ?", start, end).
		Group("user_id, subscribe_id"). // 修改这里，添加 subscribe_id 到 GROUP BY 子句
		Order("total DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

// QueryTrafficLogPageList returns a list of records that meet the conditions.
func (m *customTrafficModel) QueryTrafficLogPageList(ctx context.Context, userId, subscribeId int64, page, size int) ([]*TrafficLog, int64, error) {
	var list []*TrafficLog
	var total int64
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).Where("user_id = ? and subscribe_id= ?", userId, subscribeId).Count(&total).Limit(size).Offset((page - 1) * size).Find(&list).Error
	return list, total, err
}
