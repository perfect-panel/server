package traffic

import (
	"context"
	"fmt"
	"time"

	"github.com/perfect-panel/server/pkg/orm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type customTrafficLogicModel interface {
	QueryServerTrafficByDay(ctx context.Context, serverId int64, date time.Time) (*TotalTraffic, error)
	QueryTrafficByDay(ctx context.Context, date time.Time) (*TotalTraffic, error)
	QueryTrafficByMonthly(ctx context.Context, date time.Time) (*TotalTraffic, error)
	QueryTrafficSummary(ctx context.Context, start, end time.Time) (*TotalTraffic, error)
	TopServersTrafficByDay(ctx context.Context, date time.Time, limit int) ([]ServerTrafficRanking, error)
	TopServersTrafficByMonthly(ctx context.Context, date time.Time, limit int) ([]ServerTrafficRanking, error)
	TopUsersTrafficByDay(ctx context.Context, date time.Time, limit int) ([]UserTrafficRanking, error)
	TopUsersTrafficByMonthly(ctx context.Context, date time.Time, limit int) ([]UserTrafficRanking, error)
	QueryServerTrafficRanking(ctx context.Context, start, end time.Time) ([]ServerTrafficRanking, error)
	QueryUserTrafficRanking(ctx context.Context, start, end time.Time) ([]UserTrafficRanking, error)
	QueryTrafficLogPageList(ctx context.Context, userId, subscribeId int64, page, size int) ([]*TrafficLog, int64, error)
	QueryTrafficLogDetails(ctx context.Context, filter *TrafficLogDetailsFilter) ([]*TrafficLog, int64, error)
	DeleteBefore(ctx context.Context, end time.Time) error
}

type TrafficLogDetailsFilter struct {
	ServerId    int64
	UserId      int64
	SubscribeId int64
	Start       time.Time
	End         time.Time
	Page        int
	Size        int
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB) Model {
	return &customTrafficModel{
		defaultTrafficModel: newTrafficModel(conn),
	}
}

func (m *customTrafficModel) QueryServerTrafficByDay(ctx context.Context, serverId int64, date time.Time) (*TotalTraffic, error) {
	var data TotalTraffic
	start, end := dayRange(date)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(totalTrafficSelect(m.Conn)).
		Where(fmt.Sprintf("%s = ? AND %s >= ? AND %s < ?", trafficColumn(m.Conn, "server_id"), trafficColumn(m.Conn, "timestamp"), trafficColumn(m.Conn, "timestamp")), serverId, start, end).
		Scan(&data).Error
	return &data, err
}

func (m *customTrafficModel) QueryTrafficByDay(ctx context.Context, date time.Time) (*TotalTraffic, error) {
	var data TotalTraffic
	start, end := dayRange(date)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(totalTrafficSelect(m.Conn)).
		Where(timeRangeCondition(m.Conn), start, end).
		Scan(&data).Error
	return &data, err
}

func (m *customTrafficModel) QueryTrafficByMonthly(ctx context.Context, date time.Time) (*TotalTraffic, error) {
	var data TotalTraffic
	start, end := monthRange(date)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(totalTrafficSelect(m.Conn)).
		Where(timeRangeCondition(m.Conn), start, end).
		Scan(&data).Error
	return &data, err
}

func (m *customTrafficModel) QueryTrafficSummary(ctx context.Context, start, end time.Time) (*TotalTraffic, error) {
	var data TotalTraffic
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(totalTrafficSelect(m.Conn)).
		Where(timeRangeCondition(m.Conn), start, end).
		Scan(&data).Error
	return &data, err
}

func (m *customTrafficModel) TopServersTrafficByDay(ctx context.Context, date time.Time, limit int) ([]ServerTrafficRanking, error) {
	var summaries []ServerTrafficRanking
	start, end := dayRange(date)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(serverTrafficRankingSelect(m.Conn)).
		Where(timeRangeCondition(m.Conn), start, end).
		Group(trafficColumn(m.Conn, "server_id")).
		Order("total DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

func (m *customTrafficModel) TopServersTrafficByMonthly(ctx context.Context, date time.Time, limit int) ([]ServerTrafficRanking, error) {
	var summaries []ServerTrafficRanking
	start, end := monthRange(date)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(serverTrafficRankingSelect(m.Conn)).
		Where(timeRangeCondition(m.Conn), start, end).
		Group(trafficColumn(m.Conn, "server_id")).
		Order("total DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

func (m *customTrafficModel) TopUsersTrafficByDay(ctx context.Context, date time.Time, limit int) ([]UserTrafficRanking, error) {
	var summaries []UserTrafficRanking
	start, end := dayRange(date)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(userTrafficRankingSelect(m.Conn)).
		Where(timeRangeCondition(m.Conn), start, end).
		Group(trafficColumn(m.Conn, "user_id") + ", " + trafficColumn(m.Conn, "subscribe_id")).
		Order("total DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

func (m *customTrafficModel) TopUsersTrafficByMonthly(ctx context.Context, date time.Time, limit int) ([]UserTrafficRanking, error) {
	var summaries []UserTrafficRanking
	start, end := monthRange(date)
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(userTrafficRankingSelect(m.Conn)).
		Where(timeRangeCondition(m.Conn), start, end).
		Group(trafficColumn(m.Conn, "user_id") + ", " + trafficColumn(m.Conn, "subscribe_id")).
		Order("total DESC").
		Limit(limit).
		Scan(&summaries).Error
	return summaries, err
}

func (m *customTrafficModel) QueryServerTrafficRanking(ctx context.Context, start, end time.Time) ([]ServerTrafficRanking, error) {
	var summaries []ServerTrafficRanking
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(serverTrafficRankingSelect(m.Conn)).
		Where(timeRangeCondition(m.Conn), start, end).
		Group(trafficColumn(m.Conn, "server_id")).
		Order("total DESC").
		Scan(&summaries).Error
	return summaries, err
}

func (m *customTrafficModel) QueryUserTrafficRanking(ctx context.Context, start, end time.Time) ([]UserTrafficRanking, error) {
	var summaries []UserTrafficRanking
	err := m.Conn.WithContext(ctx).Model(&TrafficLog{}).
		Select(userTrafficRankingSelect(m.Conn)).
		Where(timeRangeCondition(m.Conn), start, end).
		Group(trafficColumn(m.Conn, "user_id") + ", " + trafficColumn(m.Conn, "subscribe_id")).
		Order("total DESC").
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

func (m *customTrafficModel) QueryTrafficLogDetails(ctx context.Context, filter *TrafficLogDetailsFilter) ([]*TrafficLog, int64, error) {
	if filter == nil {
		filter = &TrafficLogDetailsFilter{Page: 1, Size: 10}
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Size < 1 {
		filter.Size = 10
	}

	query := m.Conn.WithContext(ctx).Model(&TrafficLog{})
	if filter.ServerId != 0 {
		query = query.Where("server_id = ?", filter.ServerId)
	}
	if !filter.Start.IsZero() && !filter.End.IsZero() {
		query = query.Where(timeRangeCondition(m.Conn), filter.Start, filter.End)
	}
	if filter.UserId != 0 {
		query = query.Where("user_id = ?", filter.UserId)
	}
	if filter.SubscribeId != 0 {
		query = query.Where("subscribe_id = ?", filter.SubscribeId)
	}

	var list []*TrafficLog
	var total int64
	err := query.Count(&total).
		Order("timestamp DESC").
		Limit(filter.Size).
		Offset((filter.Page - 1) * filter.Size).
		Find(&list).Error
	return list, total, err
}

func (m *customTrafficModel) DeleteBefore(ctx context.Context, end time.Time) error {
	return m.Conn.WithContext(ctx).Model(&TrafficLog{}).Where(trafficColumn(m.Conn, "timestamp")+" <= ?", end).Delete(&TrafficLog{}).Error
}

func dayRange(date time.Time) (time.Time, time.Time) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return start, start.Add(24 * time.Hour)
}

func monthRange(date time.Time) (time.Time, time.Time) {
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	return start, start.AddDate(0, 1, 0)
}

func timeRangeCondition(db *gorm.DB) string {
	column := trafficColumn(db, "timestamp")
	return column + " >= ? AND " + column + " < ?"
}

func totalTrafficSelect(db *gorm.DB) string {
	return sumIntExpr(db, trafficColumn(db, "download"), "download") + ", " +
		sumIntExpr(db, trafficColumn(db, "upload"), "upload")
}

func serverTrafficRankingSelect(db *gorm.DB) string {
	download := trafficColumn(db, "download")
	upload := trafficColumn(db, "upload")
	return fmt.Sprintf(
		"%s AS server_id, %s, %s, %s",
		trafficColumn(db, "server_id"),
		sumIntExpr(db, download+" + "+upload, "total"),
		sumIntExpr(db, download, "download"),
		sumIntExpr(db, upload, "upload"),
	)
}

func userTrafficRankingSelect(db *gorm.DB) string {
	download := trafficColumn(db, "download")
	upload := trafficColumn(db, "upload")
	return fmt.Sprintf(
		"%s AS user_id, %s AS subscribe_id, %s, %s, %s",
		trafficColumn(db, "user_id"),
		trafficColumn(db, "subscribe_id"),
		sumIntExpr(db, download+" + "+upload, "total"),
		sumIntExpr(db, download, "download"),
		sumIntExpr(db, upload, "upload"),
	)
}

func sumIntExpr(db *gorm.DB, expr, alias string) string {
	if db != nil && db.Dialector.Name() == orm.DriverPostgres {
		return fmt.Sprintf("COALESCE(SUM(%s), 0)::bigint AS %s", expr, alias)
	}
	return fmt.Sprintf("COALESCE(SUM(%s), 0) AS %s", expr, alias)
}

func trafficColumn(db *gorm.DB, column string) string {
	if db != nil && db.Statement != nil {
		return db.Statement.Quote(clause.Column{Table: TrafficLog{}.TableName(), Name: column})
	}
	return TrafficLog{}.TableName() + "." + column
}
