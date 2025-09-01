package console

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type QueryServerTotalDataLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewQueryServerTotalDataLogic Query server total data
func NewQueryServerTotalDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryServerTotalDataLogic {
	return &QueryServerTotalDataLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryServerTotalDataLogic) QueryServerTotalData() (resp *types.ServerTotalDataResponse, err error) {

	if strings.ToLower(os.Getenv("PPANEL_MODE")) == "demo" {
		return l.mockRevenueStatistics(), nil
	}

	now := time.Now()

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour).Add(-time.Second)
	query := l.svcCtx.DB.WithContext(l.ctx)
	var todayTop10User []log.UserTraffic

	err = query.Model(&traffic.TrafficLog{}).
		Select("user_id, subscribe_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
		Where("timestamp BETWEEN ? AND ?", todayStart, todayEnd).
		Group("user_id, subscribe_id").
		Order("total DESC").
		Limit(10).
		Scan(&todayTop10User).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("[Traffic Stat Queue] Query user traffic failed: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), " Query user traffic failed: %v", err.Error())
	}
	var userTodayTrafficRanking []types.UserTrafficData
	for _, item := range todayTop10User {
		userTodayTrafficRanking = append(userTodayTrafficRanking, types.UserTrafficData{
			SID:      item.SubscribeId,
			Upload:   item.Upload,
			Download: item.Download,
		})
	}

	// query yesterday user traffic rank log
	yesterday := todayStart.Add(-24 * time.Hour).Format(time.DateOnly)

	var yesterdayLog log.SystemLog
	err = query.Model(&log.SystemLog{}).Where("`date` = ? AND `type` = ?", yesterday, log.TypeUserTrafficRank).First(&yesterdayLog).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorw("[QueryServerTotalDataLogic] Query yesterday user traffic rank log error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Query yesterday user traffic rank log error: %v", err)
	}

	var yesterdayUserRankData []types.UserTrafficData
	if yesterdayLog.Id > 0 {
		var rank log.UserTrafficRank
		err = rank.Unmarshal([]byte(yesterdayLog.Content))
		if err != nil {
			l.Errorw("[QueryServerTotalDataLogic] Unmarshal yesterday user traffic rank log error", logger.Field("error", err.Error()))
		}
		yesterdayUserRankData = make([]types.UserTrafficData, 0, len(rank.Rank))
		for i, v := range rank.Rank {
			yesterdayUserRankData[i] = types.UserTrafficData{
				SID:      v.SubscribeId,
				Upload:   v.Upload,
				Download: v.Download,
			}
		}
	}

	// query server traffic rank today
	var todayTop10Server []log.ServerTraffic
	err = query.Model(&traffic.TrafficLog{}).Select("server_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
		Where("timestamp BETWEEN ? AND ?", todayStart, todayEnd).
		Group("server_id").
		Order("total DESC").
		Limit(10).
		Scan(&todayTop10Server).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("[Traffic Stat Queue] Query server traffic failed: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), " Query server traffic failed: %v", err.Error())
	}

	var todayServerRanking []types.ServerTrafficData
	for _, item := range todayTop10Server {
		info, err := l.svcCtx.NodeModel.FindOneServer(l.ctx, item.ServerId)
		if err != nil {
			l.Errorw("[QueryServerTotalDataLogic] FindOneServer error", logger.Field("error", err.Error()), logger.Field("server_id", item.ServerId))
			continue
		}
		todayServerRanking = append(todayServerRanking, types.ServerTrafficData{
			ServerId: item.ServerId,
			Name:     info.Name,
			Upload:   item.Upload,
			Download: item.Download,
		})
	}

	// query server traffic rank yesterday
	var yesterdayTop10Server []types.ServerTrafficData
	var yesterdayServerTrafficLog log.SystemLog
	err = query.Model(&log.SystemLog{}).Where("`date` = ? AND `type` = ?", yesterday, log.TypeServerTrafficRank).First(&yesterdayServerTrafficLog).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorw("[QueryServerTotalDataLogic] Query yesterday server traffic rank log error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Query yesterday server traffic rank log error: %v", err)
	}
	if yesterdayServerTrafficLog.Id > 0 {
		var rank log.ServerTrafficRank
		err = rank.Unmarshal([]byte(yesterdayServerTrafficLog.Content))
		if err != nil {
			l.Errorw("[QueryServerTotalDataLogic] Unmarshal yesterday server traffic rank log error", logger.Field("error", err.Error()))
		}
		yesterdayTop10Server = make([]types.ServerTrafficData, 0, len(rank.Rank))
		for i, v := range rank.Rank {
			info, err := l.svcCtx.NodeModel.FindOneServer(l.ctx, v.ServerId)
			if err != nil {
				l.Errorw("[QueryServerTotalDataLogic] FindOneServer error", logger.Field("error", err.Error()), logger.Field("server_id", v.ServerId))
				continue
			}
			yesterdayTop10Server[i] = types.ServerTrafficData{
				ServerId: v.ServerId,
				Name:     info.Name,
				Upload:   v.Upload,
				Download: v.Download,
			}
		}
	}

	resp = &types.ServerTotalDataResponse{
		ServerTrafficRankingToday:     todayServerRanking,
		ServerTrafficRankingYesterday: yesterdayTop10Server,
		UserTrafficRankingToday:       userTodayTrafficRanking,
		UserTrafficRankingYesterday:   yesterdayUserRankData,
	}

	return resp, nil
}

func (l *QueryServerTotalDataLogic) mockRevenueStatistics() *types.ServerTotalDataResponse {
	now := time.Now()

	// Generate server traffic ranking data for today (top 10)
	serverTrafficToday := make([]types.ServerTrafficData, 10)
	serverNames := []string{"香港-01", "美国-洛杉矶", "日本-东京", "新加坡-01", "韩国-首尔", "台湾-01", "德国-法兰克福", "英国-伦敦", "加拿大-多伦多", "澳洲-悉尼"}
	for i := 0; i < 10; i++ {
		upload := int64(500000000 + (i * 100000000) + (i%3)*200000000)    // 500MB - 1.5GB
		download := int64(2000000000 + (i * 300000000) + (i%4)*500000000) // 2GB - 8GB
		serverTrafficToday[i] = types.ServerTrafficData{
			ServerId: int64(i + 1),
			Name:     serverNames[i],
			Upload:   upload,
			Download: download,
		}
	}

	// Generate server traffic ranking data for yesterday (top 10)
	serverTrafficYesterday := make([]types.ServerTrafficData, 10)
	for i := 0; i < 10; i++ {
		upload := int64(480000000 + (i * 95000000) + (i%3)*180000000)
		download := int64(1900000000 + (i * 280000000) + (i%4)*450000000)
		serverTrafficYesterday[i] = types.ServerTrafficData{
			ServerId: int64(i + 1),
			Name:     serverNames[i],
			Upload:   upload,
			Download: download,
		}
	}

	//// Generate user traffic ranking data for today (top 10)
	//userTrafficToday := make([]types.UserTrafficData, 10)
	//for i := 0; i < 10; i++ {
	//	upload := int64(100000000 + (i*20000000) + (i%5)*50000000)   // 100MB - 400MB
	//	download := int64(800000000 + (i*150000000) + (i%3)*300000000) // 800MB - 3GB
	//	userTrafficToday[i] = types.UserTrafficData{
	//		SID:      int64(10001 + i),
	//		Upload:   upload,
	//		Download: download,
	//	}
	//}

	//// Generate user traffic ranking data for yesterday (top 10)
	//userTrafficYesterday := make([]types.UserTrafficData, 10)
	//for i := 0; i < 10; i++ {
	//	upload := int64(95000000 + (i*18000000) + (i%5)*45000000)
	//	download := int64(750000000 + (i*140000000) + (i%3)*280000000)
	//	userTrafficYesterday[i] = types.UserTrafficData{
	//		SID:      int64(10001 + i),
	//		Upload:   upload,
	//		Download: download,
	//	}
	//}
	//
	return &types.ServerTotalDataResponse{
		OnlineUserIPs:                 1688,
		OnlineServers:                 8,
		OfflineServers:                2,
		TodayUpload:                   8888888888,   // ~8.3GB
		TodayDownload:                 28888888888,  // ~26.9GB
		MonthlyUpload:                 288888888888, // ~269GB
		MonthlyDownload:               888888888888, // ~828GB
		UpdatedAt:                     now.Unix(),
		ServerTrafficRankingToday:     serverTrafficToday,
		ServerTrafficRankingYesterday: serverTrafficYesterday,
		//UserTrafficRankingToday:       userTrafficToday,
		//UserTrafficRankingYesterday:   userTrafficYesterday,
	}
}
