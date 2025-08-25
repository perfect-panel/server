package console

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
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

	resp = &types.ServerTotalDataResponse{
		ServerTrafficRankingToday:     make([]types.ServerTrafficData, 0),
		ServerTrafficRankingYesterday: make([]types.ServerTrafficData, 0),
		UserTrafficRankingToday:       make([]types.UserTrafficData, 0),
		UserTrafficRankingYesterday:   make([]types.UserTrafficData, 0),
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
