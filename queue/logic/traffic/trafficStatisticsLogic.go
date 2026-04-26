package traffic

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/queue/types"
	"gorm.io/gorm"
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
	// query protocol ratio
	// default ratio is 1.0

	protocols, err := serverInfo.UnmarshalProtocols()
	if err != nil {
		logger.Errorf("[TrafficStatistics] Unmarshal protocols failed: %s", err.Error())
		return nil
	}
	var protocol *node.Protocol

	var ratio float32 = 1.0

	for _, p := range protocols {
		if strings.ToLower(p.Type) == strings.ToLower(payload.Protocol) {
			protocol = &p
			break
		}
	}

	if protocol == nil {
		logger.WithContext(ctx).Error("[TrafficStatistics] Protocol not found: %s", payload.Protocol)
		return nil
	}

	// use protocol ratio if it's greater than 0
	if protocol.Ratio > 0 {
		ratio = float32(protocol.Ratio)
	}

	now := time.Now()
	realTimeMultiplier := l.svc.NodeMultiplierManager.GetMultiplier(now)
	logger.Debugf("[TrafficStatisticsLogic] Current time traffic multiplier: %.2f", realTimeMultiplier)

	// V4.3 决策 32:节点上报的 SID 现在是 device.id(getServerUserList 输出粒度)。
	// 需要先把 device → user_subscribe_id 解析,然后按 user_subscribe_id 聚合写入。
	// 兼容老节点:若 SID 找不到 device,fall back 到把它当 user_subscribe.id 处理。
	type aggKey struct {
		userSubscribeId int64
		userId          int64
	}
	type aggVal struct {
		download int64
		upload   int64
		// 设备粒度的明细,用于 today_traffic + last_seen_at 写回
		devices map[int64]*types.UserTraffic
	}
	agg := make(map[aggKey]*aggVal, len(payload.Logs))

	for _, lg := range payload.Logs {
		if lg.Download+lg.Upload <= l.svc.Config.Node.TrafficReportThreshold {
			continue
		}
		var userSubId, userId, deviceId int64
		// 优先按 device 解析(V4.3 默认路径)
		dev, err := l.svc.UserModel.FindOneSubscribeDevice(ctx, lg.SID)
		switch {
		case err == nil:
			userSubId = dev.UserSubscribeId
			userId = dev.UserId
			deviceId = dev.Id
		case errors.Is(err, gorm.ErrRecordNotFound):
			// fallback:legacy/老数据,SID == user_subscribe.id
			sub, err2 := l.svc.UserModel.FindOneSubscribe(ctx, lg.SID)
			if err2 != nil {
				logger.WithContext(ctx).Error("[TrafficStatistics] resolve SID failed",
					logger.Field("sid", lg.SID),
					logger.Field("device_err", err.Error()),
					logger.Field("subscribe_err", err2.Error()),
				)
				continue
			}
			userSubId = sub.Id
			userId = sub.UserId
		default:
			logger.WithContext(ctx).Error("[TrafficStatistics] device lookup failed",
				logger.Field("sid", lg.SID), logger.Field("error", err.Error()))
			continue
		}

		d := int64(float32(lg.Download) * ratio * realTimeMultiplier)
		u := int64(float32(lg.Upload) * ratio * realTimeMultiplier)

		key := aggKey{userSubscribeId: userSubId, userId: userId}
		v, ok := agg[key]
		if !ok {
			v = &aggVal{devices: map[int64]*types.UserTraffic{}}
			agg[key] = v
		}
		v.download += d
		v.upload += u
		if deviceId > 0 {
			v.devices[deviceId] = &types.UserTraffic{SID: deviceId, Download: d, Upload: u}
		}
	}

	for key, v := range agg {
		if err := l.svc.UserModel.UpdateUserSubscribeWithTraffic(ctx, key.userSubscribeId, v.download, v.upload); err != nil {
			logger.WithContext(ctx).Error("[TrafficStatistics] aggregate write failed",
				logger.Field("user_subscribe_id", key.userSubscribeId),
				logger.Field("error", err.Error()),
			)
			continue
		}
		// V4.3 决策 6 + 40:实时检测是否突破 100% 配额,首次越线设 throttled_at + cut_off_at,
		// DEL Redis user list 缓存让节点下一次拉取时收到 speed_limit=1 Mbps。
		l.checkAndTriggerThrottle(ctx, key.userSubscribeId)
		// 设备级 last_seen_at + today_traffic 累加(轻量,失败不阻断主流)
		for devId, dt := range v.devices {
			dev, err := l.svc.UserModel.FindOneSubscribeDevice(ctx, devId)
			if err != nil {
				continue
			}
			t := now
			dev.LastSeenAt = &t
			dev.TodayTraffic += dt.Download + dt.Upload
			_ = l.svc.UserModel.UpdateSubscribeDevice(ctx, dev)
		}
		// 流量明细日志(沿用 SubscribeId = user_subscribe_id)
		if err := l.svc.TrafficLogModel.Insert(ctx, &traffic.TrafficLog{
			ServerId:    payload.ServerId,
			SubscribeId: key.userSubscribeId,
			UserId:      key.userId,
			Upload:      v.upload,
			Download:    v.download,
			Timestamp:   now,
		}); err != nil {
			logger.WithContext(ctx).Error("[TrafficStatistics] write traffic log failed",
				logger.Field("user_subscribe_id", key.userSubscribeId),
				logger.Field("error", err.Error()),
			)
		}
	}
	return nil
}

// checkAndTriggerThrottle — 100% 配额触发限速(决策 6+40)。
// - 仅在首次越线时设 throttled_at + cut_off_at,避免重复写。
// - DEL `server:user:*` 让节点下一次 pull 收到 speed_limit=1 Mbps。
// - 通知由 Phase 6 cron 异步发送(模板 throttle_started)。
func (l *TrafficStatisticsLogic) checkAndTriggerThrottle(ctx context.Context, userSubscribeId int64) {
	sub, err := l.svc.UserModel.FindOneSubscribe(ctx, userSubscribeId)
	if err != nil {
		return
	}
	if sub.ThrottledAt != nil {
		return // 已限速
	}
	quota := sub.Traffic + sub.TrafficAddon
	if quota <= 0 {
		return
	}
	if sub.Download+sub.Upload < quota {
		return
	}
	now := time.Now()
	cut := now.Add(24 * time.Hour)
	sub.ThrottledAt = &now
	sub.CutOffAt = &cut
	sub.Notified100 = false
	sub.Notified12h = false
	sub.Notified24h = false
	if err := l.svc.UserModel.UpdateSubscribe(ctx, sub); err != nil {
		logger.WithContext(ctx).Error("[checkAndTriggerThrottle] update failed",
			logger.Field("user_subscribe_id", userSubscribeId),
			logger.Field("error", err.Error()))
		return
	}
	// 失活节点缓存 — 与 reset 流程共用同一 helper(Phase 4)
	pattern := "server:user:*"
	iter := l.svc.Redis.Scan(ctx, 0, pattern, 100).Iterator()
	keys := make([]string, 0, 16)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		_ = l.svc.Redis.Del(ctx, keys...).Err()
	}
}
