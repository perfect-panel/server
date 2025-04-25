package console

import (
	"context"
	"time"

	"github.com/perfect-panel/server/pkg/xerr"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/pkg/errors"
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
	resp = &types.ServerTotalDataResponse{
		ServerTrafficRankingToday:     make([]types.ServerTrafficData, 0),
		ServerTrafficRankingYesterday: make([]types.ServerTrafficData, 0),
		UserTrafficRankingToday:       make([]types.UserTrafficData, 0),
		UserTrafficRankingYesterday:   make([]types.UserTrafficData, 0),
	}

	// Query node server status
	servers, err := l.svcCtx.ServerModel.FindAllServer(l.ctx)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] FindAllServer error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(err, "FindAllServer error: %v", err)
	}
	onlineServers, err := l.svcCtx.NodeCache.GetOnlineNodeStatusCount(l.ctx)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] GetOnlineNodeStatusCount error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(err, "GetOnlineNodeStatusCount error: %v", err)
	}
	resp.OnlineServers = onlineServers
	resp.OfflineServers = int64(len(servers) - int(onlineServers))

	// 获取所有节点在线用户
	allNodeOnlineUser, err := l.svcCtx.NodeCache.GetAllNodeOnlineUser(l.ctx)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] Get all node online user failed", logger.Field("error", err.Error()))
	}
	resp.OnlineUserIPs = int64(len(allNodeOnlineUser))

	// 获取所有节点今日上传下载流量
	allNodeUploadTraffic, err := l.svcCtx.NodeCache.GetAllNodeUploadTraffic(l.ctx)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] Get all node upload traffic failed", logger.Field("error", err.Error()))
	}
	resp.TodayUpload = allNodeUploadTraffic
	allNodeDownloadTraffic, err := l.svcCtx.NodeCache.GetAllNodeDownloadTraffic(l.ctx)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] Get all node download traffic failed", logger.Field("error", err.Error()))
	}
	resp.TodayDownload = allNodeDownloadTraffic
	// 获取节点流量排行榜 前10
	nodeTrafficRankingToday, err := l.svcCtx.NodeCache.GetNodeTodayTotalTrafficRank(l.ctx, 10)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] Get node today total traffic rank failed", logger.Field("error", err.Error()))
	}
	if len(nodeTrafficRankingToday) > 0 {
		var serverTrafficData []types.ServerTrafficData
		for _, rank := range nodeTrafficRankingToday {
			serverInfo, err := l.svcCtx.ServerModel.FindOne(l.ctx, rank.ID)
			if err != nil {
				l.Errorw("[QueryServerTotalDataLogic] FindOne error", logger.Field("error", err))
				continue
			}
			serverTrafficData = append(serverTrafficData, types.ServerTrafficData{
				ServerId: rank.ID,
				Name:     serverInfo.Name,
				Upload:   rank.Upload,
				Download: rank.Download,
			})
		}
		resp.ServerTrafficRankingToday = serverTrafficData
	}
	// 获取用户流量排行榜 前10
	userTrafficRankingToday, err := l.svcCtx.NodeCache.GetUserTodayTotalTrafficRank(l.ctx, 10)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] Get user today total traffic rank failed", logger.Field("error", err.Error()))
	}

	if len(userTrafficRankingToday) > 0 {
		var userTrafficData []types.UserTrafficData
		for _, rank := range userTrafficRankingToday {
			userTrafficData = append(userTrafficData, types.UserTrafficData{
				SID:      rank.SID,
				Upload:   rank.Upload,
				Download: rank.Download,
			})
		}
		resp.UserTrafficRankingToday = userTrafficData
	}
	// 获取昨日节点流量排行榜 前10
	nodeTrafficRankingYesterday, err := l.svcCtx.NodeCache.GetYesterdayNodeTotalTrafficRank(l.ctx)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] Get yesterday node total traffic rank failed", logger.Field("error", err.Error()))
	}
	if len(nodeTrafficRankingYesterday) > 0 {
		var serverTrafficData []types.ServerTrafficData
		for _, rank := range nodeTrafficRankingYesterday {
			serverTrafficData = append(serverTrafficData, types.ServerTrafficData{
				ServerId: rank.ID,
				Name:     rank.Name,
				Upload:   rank.Upload,
				Download: rank.Download,
			})
		}
		resp.ServerTrafficRankingYesterday = serverTrafficData
	}
	// 获取昨日用户流量排行榜 前10
	userTrafficRankingYesterday, err := l.svcCtx.NodeCache.GetYesterdayUserTotalTrafficRank(l.ctx)
	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] Get yesterday user total traffic rank failed", logger.Field("error", err.Error()))
	}
	if len(userTrafficRankingYesterday) > 0 {
		var userTrafficData []types.UserTrafficData
		for _, rank := range userTrafficRankingYesterday {
			userTrafficData = append(userTrafficData, types.UserTrafficData{
				SID:      rank.SID,
				Upload:   rank.Upload,
				Download: rank.Download,
			})
		}
		resp.UserTrafficRankingYesterday = userTrafficData
	}

	// Query node traffic by monthly
	nodeTraffic, err := l.svcCtx.TrafficLogModel.QueryTrafficByMonthly(l.ctx, time.Now())

	if err != nil {
		l.Errorw("[QueryServerTotalDataLogic] QueryTrafficByMonthly error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "QueryTrafficByMonthly error: %v", err.Error())
	}
	resp.MonthlyUpload = nodeTraffic.Upload
	resp.MonthlyDownload = nodeTraffic.Download

	return resp, nil
}
