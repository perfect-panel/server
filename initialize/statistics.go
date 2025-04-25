package initialize

import (
	"context"
	"time"

	"github.com/perfect-panel/server/internal/model/cache"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

func TrafficDataToRedis(svcCtx *svc.ServiceContext) {
	ctx := context.Background()
	// 统计昨天的节点流量数据排行榜前10
	nodeData, err := svcCtx.TrafficLogModel.TopServersTrafficByDay(ctx, time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 0, 0, 0, 0, time.Local), 10)
	if err != nil {
		logger.Errorw("统计昨天的流量数据失败", logger.Field("error", err.Error()))
	}
	var nodeCacheData []cache.NodeTodayTrafficRank
	for _, node := range nodeData {
		serverInfo, err := svcCtx.ServerModel.FindOne(ctx, node.ServerId)
		if err != nil {
			logger.Errorw("查询节点信息失败", logger.Field("error", err.Error()))
			continue
		}
		nodeCacheData = append(nodeCacheData, cache.NodeTodayTrafficRank{
			ID:       node.ServerId,
			Name:     serverInfo.Name,
			Upload:   node.Upload,
			Download: node.Download,
			Total:    node.Upload + node.Download,
		})
	}
	// 写入缓存
	if err = svcCtx.NodeCache.UpdateYesterdayNodeTotalTrafficRank(ctx, nodeCacheData); err != nil {
		logger.Errorw("写入昨天的流量数据到缓存失败", logger.Field("error", err.Error()))
	}
	// 统计昨天的用户流量数据排行榜前10
	userData, err := svcCtx.TrafficLogModel.TopUsersTrafficByDay(ctx, time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 0, 0, 0, 0, time.Local), 10)
	if err != nil {
		logger.Errorw("统计昨天的流量数据失败", logger.Field("error", err.Error()))
	}
	var userCacheData []cache.UserTodayTrafficRank
	for _, user := range userData {
		userCacheData = append(userCacheData, cache.UserTodayTrafficRank{
			SID:      user.SubscribeId,
			Upload:   user.Upload,
			Download: user.Download,
			Total:    user.Upload + user.Download,
		})
	}
	// 写入缓存
	if err = svcCtx.NodeCache.UpdateYesterdayUserTotalTrafficRank(ctx, userCacheData); err != nil {
		logger.Errorw("写入昨天的流量数据到缓存失败", logger.Field("error", err.Error()))
	}
	logger.Infow("初始化昨天的流量数据到缓存成功")
}
