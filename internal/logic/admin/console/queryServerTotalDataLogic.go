package console

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const consoleServerTotalDataCacheKey = "console:server_total_data"
const consoleServerTotalDataCacheTTL = 60 * time.Second

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

	// Try cache first
	cached, cacheErr := l.svcCtx.Redis.Get(l.ctx, consoleServerTotalDataCacheKey).Result()
	if cacheErr == nil && cached != "" {
		var result types.ServerTotalDataResponse
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	query := l.svcCtx.DB.WithContext(l.ctx)

	// Parallelize three traffic_log queries to reduce latency
	var (
		todayTop10User   []log.UserTraffic
		todayTop10Server []log.ServerTraffic
		todayTraffic     struct {
			Upload   int64
			Download int64
		}
		userErr, serverErr, trafficErr error
		wg                             sync.WaitGroup
	)

	wg.Add(3)

	// Query 1: Today's top 10 users by traffic
	go func() {
		defer wg.Done()
		userErr = query.Model(&traffic.TrafficLog{}).
			Select("user_id, subscribe_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
			Where("timestamp >= ? AND timestamp < ?", todayStart, todayEnd).
			Group("user_id, subscribe_id").
			Order("total DESC").
			Limit(10).
			Scan(&todayTop10User).Error
	}()

	// Query 2: Today's top 10 servers by traffic
	go func() {
		defer wg.Done()
		serverErr = query.Model(&traffic.TrafficLog{}).
			Select("server_id, SUM(download + upload) AS total, SUM(download) AS download, SUM(upload) AS upload").
			Where("timestamp >= ? AND timestamp < ?", todayStart, todayEnd).
			Group("server_id").
			Order("total DESC").
			Limit(10).
			Scan(&todayTop10Server).Error
	}()

	// Query 3: Today's total upload/download
	go func() {
		defer wg.Done()
		trafficErr = query.Model(&traffic.TrafficLog{}).
			Select("COALESCE(SUM(upload), 0) AS upload, COALESCE(SUM(download), 0) AS download").
			Where("timestamp >= ? AND timestamp < ?", todayStart, todayEnd).
			Scan(&todayTraffic).Error
	}()

	wg.Wait()

	if userErr != nil && !errors.Is(userErr, gorm.ErrRecordNotFound) {
		logger.Errorf("[QueryServerTotalData] Query user traffic failed: %v", userErr.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Query user traffic failed: %v", userErr.Error())
	}
	if serverErr != nil && !errors.Is(serverErr, gorm.ErrRecordNotFound) {
		logger.Errorf("[QueryServerTotalData] Query server traffic failed: %v", serverErr.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Query server traffic failed: %v", serverErr.Error())
	}
	if trafficErr != nil && !errors.Is(trafficErr, gorm.ErrRecordNotFound) {
		logger.Errorf("[QueryServerTotalData] Sum today traffic failed: %v", trafficErr.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Sum today traffic failed: %v", trafficErr.Error())
	}

	// Build today user traffic ranking
	var userTodayTrafficRanking []types.UserTrafficData
	for _, item := range todayTop10User {
		userTodayTrafficRanking = append(userTodayTrafficRanking, types.UserTrafficData{
			SID:      item.SubscribeId,
			Upload:   item.Upload,
			Download: item.Download,
		})
	}

	// Query yesterday user traffic rank log
	yesterday := todayStart.Add(-24 * time.Hour).Format(time.DateOnly)

	var yesterdayLog log.SystemLog
	err = query.Model(&log.SystemLog{}).Where("date = ? AND type = ?", yesterday, log.TypeUserTrafficRank).First(&yesterdayLog).Error
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
		for _, v := range rank.Rank {
			yesterdayUserRankData = append(yesterdayUserRankData, types.UserTrafficData{
				SID:      v.SubscribeId,
				Upload:   v.Upload,
				Download: v.Download,
			})
		}
	}

	// Batch fetch server names for today's ranking
	serverIDs := make([]int64, 0, 10)
	for _, item := range todayTop10Server {
		serverIDs = append(serverIDs, item.ServerId)
	}

	serverMap := make(map[int64]*node.Server)
	if len(serverIDs) > 0 {
		var servers []*node.Server
		err = query.Model(&node.Server{}).Where("id IN ?", serverIDs).Find(&servers).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorw("[QueryServerTotalDataLogic] Batch fetch servers error", logger.Field("error", err.Error()))
		}
		for _, s := range servers {
			serverMap[s.Id] = s
		}
	}

	var todayServerRanking []types.ServerTrafficData
	for _, item := range todayTop10Server {
		name := ""
		if s, ok := serverMap[item.ServerId]; ok {
			name = s.Name
		}
		todayServerRanking = append(todayServerRanking, types.ServerTrafficData{
			ServerId: item.ServerId,
			Name:     name,
			Upload:   item.Upload,
			Download: item.Download,
		})
	}

	// Query yesterday server traffic rank
	var yesterdayTop10Server []types.ServerTrafficData
	var yesterdayServerTrafficLog log.SystemLog
	err = query.Model(&log.SystemLog{}).Where("date = ? AND type = ?", yesterday, log.TypeServerTrafficRank).First(&yesterdayServerTrafficLog).Error
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

		// Collect yesterday server IDs not already fetched
		yesterdayServerIDs := make([]int64, 0, len(rank.Rank))
		for _, v := range rank.Rank {
			if _, ok := serverMap[v.ServerId]; !ok {
				yesterdayServerIDs = append(yesterdayServerIDs, v.ServerId)
			}
		}
		if len(yesterdayServerIDs) > 0 {
			var extraServers []*node.Server
			if err := query.Model(&node.Server{}).Where("id IN ?", yesterdayServerIDs).Find(&extraServers).Error; err == nil {
				for _, s := range extraServers {
					serverMap[s.Id] = s
				}
			}
		}

		for _, v := range rank.Rank {
			name := ""
			if s, ok := serverMap[v.ServerId]; ok {
				name = s.Name
			}
			yesterdayTop10Server = append(yesterdayTop10Server, types.ServerTrafficData{
				ServerId: v.ServerId,
				Name:     name,
				Upload:   v.Upload,
				Download: v.Download,
			})
		}
	}

	// Query online user count
	onlineUsers, err := l.svcCtx.Store.Node().OnlineUserSubscribeGlobal(l.ctx)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] OnlineUserSubscribeGlobal error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "OnlineUserSubscribeGlobal error: %v", err)
	}

	// Query online/offline server count
	var onlineServers, offlineServers int64
	err = query.Model(&node.Server{}).Where("last_reported_at > ?", now.Add(-5*time.Minute)).Count(&onlineServers).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorw("[QueryServerTotalDataLogic] Count online servers error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Count online servers error: %v", err)
	}

	err = query.Model(&node.Server{}).Where("last_reported_at <= ? OR last_reported_at IS NULL", now.Add(-5*time.Minute)).Count(&offlineServers).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorw("[QueryServerTotalDataLogic] Count offline servers error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Count offline servers error: %v", err)
	}

	// Monthly traffic: today's real-time data + archived daily stats for previous days
	todayUpload := todayTraffic.Upload
	todayDownload := todayTraffic.Download
	var monthlyUpload, monthlyDownload int64
	monthlyUpload += todayUpload
	monthlyDownload += todayDownload

	// Batch query all previous days' traffic stats (eliminates N+1 loop)
	if now.Day() > 1 {
		dates := make([]string, 0, now.Day()-1)
		for i := 1; i < int(now.Day()); i++ {
			d := time.Date(now.Year(), now.Month(), i, 0, 0, 0, 0, now.Location()).Format(time.DateOnly)
			dates = append(dates, d)
		}

		var dailyLogs []log.SystemLog
		err = query.Model(&log.SystemLog{}).
			Where("date IN ? AND type = ?", dates, log.TypeTrafficStat).
			Find(&dailyLogs).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorw("[QueryServerTotalDataLogic] Batch query daily traffic stats error", logger.Field("error", err.Error()))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Batch query daily traffic stats error: %v", err)
		}

		for _, logInfo := range dailyLogs {
			var stat log.TrafficStat
			if err := stat.Unmarshal([]byte(logInfo.Content)); err != nil {
				l.Errorw("[QueryServerTotalDataLogic] Unmarshal daily traffic stat error", logger.Field("error", err.Error()), logger.Field("date", logInfo.Date))
				continue
			}
			monthlyUpload += stat.Upload
			monthlyDownload += stat.Download
		}
	}

	resp = &types.ServerTotalDataResponse{
		OnlineUsers:                   onlineUsers,
		OnlineServers:                 onlineServers,
		OfflineServers:                offlineServers,
		TodayUpload:                   todayUpload,
		TodayDownload:                 todayDownload,
		MonthlyUpload:                 monthlyUpload,
		MonthlyDownload:               monthlyDownload,
		UpdatedAt:                     now.Unix(),
		ServerTrafficRankingToday:     todayServerRanking,
		ServerTrafficRankingYesterday: yesterdayTop10Server,
		UserTrafficRankingToday:       userTodayTrafficRanking,
		UserTrafficRankingYesterday:   yesterdayUserRankData,
	}

	// Cache the result
	if data, marshalErr := json.Marshal(resp); marshalErr == nil {
		l.svcCtx.Redis.Set(l.ctx, consoleServerTotalDataCacheKey, data, consoleServerTotalDataCacheTTL)
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

	return &types.ServerTotalDataResponse{
		OnlineUsers:                   1688,
		OnlineServers:                 8,
		OfflineServers:                2,
		TodayUpload:                   8888888888,
		TodayDownload:                 28888888888,
		MonthlyUpload:                 288888888888,
		MonthlyDownload:               888888888888,
		UpdatedAt:                     now.Unix(),
		ServerTrafficRankingToday:     serverTrafficToday,
		ServerTrafficRankingYesterday: serverTrafficYesterday,
	}
}
