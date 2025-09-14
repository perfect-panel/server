package traffic

import (
	"context"
	"encoding/json"
	"time"

	"github.com/perfect-panel/server/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/queue/types"
)

//goland:noinspection GoNameStartsWithPackageName
type TrafficStatisticsLogic struct {
	svc *svc.ServiceContext
}

func NewTrafficStatisticsLogic(svc *svc.ServiceContext) *TrafficStatisticsLogic {
	return &TrafficStatisticsLogic{
		svc: svc,
	}
}

func (l *TrafficStatisticsLogic) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload types.TrafficStatistics
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logger.WithContext(ctx).Error("[TrafficStatistics] Unmarshal payload failed",
			logger.Field("error", err.Error()),
			logger.Field("payload", string(task.Payload())),
		)
		return nil
	}
	if len(payload.Logs) == 0 {
		logger.WithContext(ctx).Error("[TrafficStatistics] Payload is empty")
		return nil
	}
	// query server info
	serverInfo, err := l.svc.NodeModel.FindOneServer(ctx, payload.ServerId)
	if err != nil {
		logger.WithContext(ctx).Error("[TrafficStatistics] Find server info failed",
			logger.Field("serverId", payload.ServerId),
			logger.Field("error", err.Error()),
		)
		return nil
	}
	var serverRatio float32 = 1.0
	if serverInfo.Ratio > 0 {
		serverRatio = serverInfo.Ratio
	}

	now := time.Now()
	realTimeMultiplier := l.svc.NodeMultiplierManager.GetMultiplier(now)
	for _, log := range payload.Logs {
		// update user subscribe with log
		d := int64(float32(log.Download) * serverRatio * realTimeMultiplier)
		u := int64(float32(log.Upload) * serverRatio * realTimeMultiplier)
		if err := l.svc.UserModel.UpdateUserSubscribeWithTraffic(ctx, log.SID, d, u); err != nil {
			logger.WithContext(ctx).Error("[TrafficStatistics] Update user subscribe with log failed",
				logger.Field("sid", log.SID),
				logger.Field("download", float32(log.Download)*serverRatio),
				logger.Field("upload", float32(log.Upload)*serverRatio),
				logger.Field("error", err.Error()),
			)
			continue
		}
		// query user Subscribe Info
		sub, err := l.svc.UserModel.FindOneSubscribe(ctx, log.SID)
		if err != nil {
			logger.WithContext(ctx).Error("[TrafficStatistics] Find user Subscribe Info failed",
				logger.Field("uid", log.SID),
				logger.Field("error", err.Error()),
			)
			continue
		}

		// create log log
		if err := l.svc.TrafficLogModel.Insert(ctx, &traffic.TrafficLog{
			ServerId:    payload.ServerId,
			SubscribeId: log.SID,
			UserId:      sub.UserId,
			Upload:      u,
			Download:    d,
			Timestamp:   now,
		}); err != nil {
			logger.WithContext(ctx).Error("[TrafficStatistics] Create log log failed",
				logger.Field("uid", log.SID),
				logger.Field("download", float32(log.Download)*serverRatio),
				logger.Field("upload", float32(log.Upload)*serverRatio),
				logger.Field("error", err.Error()),
			)
		}
	}
	return nil
}
