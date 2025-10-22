package traffic

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

type StatLogic struct {
	svc *svc.ServiceContext
}

func NewStatLogic(svc *svc.ServiceContext) *StatLogic {
	return &StatLogic{
		svc: svc,
	}
}

func (l *StatLogic) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	now := time.Now()
	tx := l.svc.DB.Begin()
	var err error
	defer func(err error) {
		if err != nil {
			logger.Errorf("[Traffic Stat Queue] Process task failed: %v", err.Error())
			tx.Rollback()
		} else {
			logger.Infof("[Traffic Stat Queue] Process task completed successfully, consuming: %s", time.Since(now).String())
			// 提交事务
			if err = tx.Commit().Error; err != nil {
				logger.Errorf("[Traffic Stat Queue] Commit transaction failed: %v", err.Error())
			}
		}
	}(err)

	// 获取全部有效订阅
	var userTraffic []log.UserTraffic
	// 获取统计时间范围
	start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)

	// 查询用户流量统计, 按用户和订阅分组
	err = tx.WithContext(ctx).Model(&traffic.TrafficLog{}).
		Select("user_id, subscribe_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
		Where("timestamp BETWEEN ? AND ?", start, end).
		Group("user_id, subscribe_id").
		Order("total DESC").
		Scan(&userTraffic).Error
	if err != nil {
		logger.Errorf("[Traffic Stat Queue] Query user traffic failed: %v", err.Error())
		return err
	}

	date := start.Format(time.DateOnly)

	userTop10 := log.UserTrafficRank{
		Rank: make(map[uint8]log.UserTraffic),
	}

	// 更新用户流量统计
	for i, trafficData := range userTraffic {
		if i < 10 {
			userTop10.Rank[uint8(i+1)] = trafficData
		}
		// 更新用户流量统计日志
		content, _ := trafficData.Marshal()
		err = tx.WithContext(ctx).Model(&log.SystemLog{}).Create(&log.SystemLog{
			Type:     log.TypeSubscribeTraffic.Uint8(),
			Date:     date,
			ObjectID: trafficData.SubscribeId,
			Content:  string(content),
		}).Error
		if err != nil {
			logger.Errorf("[Traffic Stat Queue] Create user traffic log failed: %v", err.Error())
			return err
		}
	}

	userTop10Content, _ := userTop10.Marshal()

	// 更新用户排行榜
	err = tx.WithContext(ctx).Model(&log.SystemLog{}).Create(&log.SystemLog{
		Type:     log.TypeUserTrafficRank.Uint8(),
		Date:     date,
		ObjectID: 0, // 0表示全局用户排行榜
		Content:  string(userTop10Content),
	}).Error
	if err != nil {
		logger.Errorf("[Traffic Stat Queue] Create user traffic rank log failed: %v", err.Error())
		return err
	}

	// 统计服务器流量
	var serverTraffic []log.ServerTraffic
	err = tx.WithContext(ctx).Model(&traffic.TrafficLog{}).
		Select("server_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
		Where("timestamp BETWEEN ? AND ?", start, end).
		Group("server_id").
		Order("total DESC").
		Scan(&serverTraffic).Error
	if err != nil {
		logger.Errorf("[Traffic Stat Queue] Query server traffic failed: %v", err.Error())
		return err
	}

	serverTop10 := log.ServerTrafficRank{
		Rank: make(map[uint8]log.ServerTraffic),
	}
	for i, trafficData := range serverTraffic {
		if i < 10 {
			serverTop10.Rank[uint8(i+1)] = trafficData
		}
		// 更新服务器流量统计日志
		content, _ := trafficData.Marshal()
		err = tx.WithContext(ctx).Model(&log.SystemLog{}).Create(&log.SystemLog{
			Type:     log.TypeServerTraffic.Uint8(),
			Date:     date,
			ObjectID: trafficData.ServerId,
			Content:  string(content),
		}).Error
		if err != nil {
			logger.Errorf("[Traffic Stat Queue] Create server traffic log failed: %v", err.Error())
			return err
		}
	}
	serverTop10Content, _ := serverTop10.Marshal()
	// 更新服务器排行榜
	err = tx.WithContext(ctx).Model(&log.SystemLog{}).Create(&log.SystemLog{
		Type:     log.TypeServerTrafficRank.Uint8(),
		Date:     date,
		ObjectID: 0, // 0表示全局服务器排行榜
		Content:  string(serverTop10Content),
	}).Error
	if err != nil {
		logger.Errorf("[Traffic Stat Queue] Create server traffic rank log failed: %v", err.Error())
		return err
	}

	// traffic stat
	var stat log.TrafficStat
	err = tx.WithContext(ctx).Model(&traffic.TrafficLog{}).
		Select("SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
		Where("timestamp BETWEEN ? AND ?", start, end).
		Scan(&stat).Error
	if err != nil {
		logger.Errorf("[Traffic Stat Queue] Query traffic stat failed: %v", err.Error())
		return err
	}

	// 更新流量统计日志
	content, _ := stat.Marshal()
	err = tx.WithContext(ctx).Model(&log.SystemLog{}).Create(&log.SystemLog{
		Type:     log.TypeTrafficStat.Uint8(),
		Date:     date,
		ObjectID: 0,
		Content:  string(content),
	}).Error
	if err != nil {
		logger.Errorf("[Traffic Stat Queue] Create traffic stat log failed: %v", err.Error())
		return err
	}

	// Delete old traffic logs
	if l.svc.Config.Log.AutoClear {
		err = tx.WithContext(ctx).Model(&traffic.TrafficLog{}).Where("timestamp <= ?", end.AddDate(0, 0, int(-l.svc.Config.Log.ClearDays))).Delete(&traffic.TrafficLog{}).Error
		if err != nil {
			logger.Errorf("[Traffic Stat Queue] Delete server traffic log failed: %v", err.Error())
		}
	}
	return nil
}
