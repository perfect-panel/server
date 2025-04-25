package user

import (
	"context"
	"sort"
	"time"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
)

type GetUserOnlineTimeStatisticsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user online time total
func NewGetUserOnlineTimeStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOnlineTimeStatisticsLogic {
	return &GetUserOnlineTimeStatisticsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserOnlineTimeStatisticsLogic) GetUserOnlineTimeStatistics() (resp *types.GetUserOnlineTimeStatisticsResponse, err error) {
	u := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	//获取历史最长在线时间
	var OnlineSeconds int64
	if err := l.svcCtx.DB.Model(user.DeviceOnlineRecord{}).Where("user_id = ?", u.Id).Select("online_seconds").Order("online_seconds desc").Limit(1).Scan(&OnlineSeconds).Error; err != nil {
		l.Logger.Error(err)
	}

	//获取历史连续最长在线天数
	var DurationDays int64
	if err := l.svcCtx.DB.Model(user.DeviceOnlineRecord{}).Where("user_id = ?", u.Id).Select("duration_days").Order("duration_days desc").Limit(1).Scan(&DurationDays).Error; err != nil {
		l.Logger.Error(err)
	}

	//获取近七天在线情况
	var userOnlineRecord []user.DeviceOnlineRecord
	if err := l.svcCtx.DB.Model(&userOnlineRecord).Where("user_id = ? and created_at >= ?", u.Id, time.Now().AddDate(0, 0, -7).Format(time.DateTime)).Order("created_at desc").Find(&userOnlineRecord).Error; err != nil {
		l.Logger.Error(err)
	}

	//获取当前连续在线天数
	var currentContinuousDays int64
	if len(userOnlineRecord) > 0 {
		currentContinuousDays = userOnlineRecord[0].DurationDays
	} else {
		currentContinuousDays = 1
	}

	var dates []string
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, -i).Format(time.DateOnly)
		dates = append(dates, date)
	}

	onlineDays := make(map[string]types.WeeklyStat)
	for _, record := range userOnlineRecord {
		//获取近七天在线情况
		onlineTime := record.OnlineTime.Format(time.DateOnly)
		if weeklyStat, ok := onlineDays[onlineTime]; ok {
			weeklyStat.Hours += float64(record.OnlineSeconds)
			onlineDays[onlineTime] = weeklyStat
		} else {
			onlineDays[onlineTime] = types.WeeklyStat{
				Hours: float64(record.OnlineSeconds),
				//根据日期获取周几
				DayName: record.OnlineTime.Weekday().String(),
			}
		}
	}

	//补全不存在的日期
	for _, date := range dates {
		if _, ok := onlineDays[date]; !ok {
			onlineTime, _ := time.Parse(time.DateOnly, date)
			onlineDays[date] = types.WeeklyStat{
				DayName: onlineTime.Weekday().String(),
			}
		}
	}

	var keys []string
	for key := range onlineDays {
		keys = append(keys, key)
	}

	//排序
	sort.Strings(keys)

	var weeklyStats []types.WeeklyStat
	for index, key := range keys {
		weeklyStat := onlineDays[key]
		weeklyStat.Day = index + 1
		weeklyStat.Hours = weeklyStat.Hours / float64(3600)
		weeklyStats = append(weeklyStats, weeklyStat)
	}

	resp = &types.GetUserOnlineTimeStatisticsResponse{
		WeeklyStats: weeklyStats,
		ConnectionRecords: types.ConnectionRecords{
			CurrentContinuousDays:   currentContinuousDays,
			HistoryContinuousDays:   DurationDays,
			LongestSingleConnection: OnlineSeconds / 60,
		},
	}
	return
}
