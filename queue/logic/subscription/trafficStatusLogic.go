package subscription

// V4.3 流量限速→断网状态机 cron(每 5 分钟):
//
// [正常] ─90%─▶ 邮件预警(notified_90=1)
//   │
//   └─100%─▶ throttle_started 实时由 trafficStatisticsLogic 设置
//                 │
//                 ├─ 12h ─▶ 邮件提醒"剩 12h 断网"(notified_12h=1)
//                 │
//                 └─ 24h ─▶ 邮件"已断网"(notified_24h=1) + DEL 节点缓存
//                              ↓
//                          [断网](节点 user list 不返回该订阅设备)
//
// 加购流量包/续费时由对应 logic 重置 throttled_at + cut_off_at + notified_*。

import (
	"context"
	"time"

	"github.com/hibiken/asynq"

	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

type TrafficStatusLogic struct {
	svc *svc.ServiceContext
}

func NewTrafficStatusLogic(svc *svc.ServiceContext) *TrafficStatusLogic {
	return &TrafficStatusLogic{svc: svc}
}

func (l *TrafficStatusLogic) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	subs, err := l.activeSubscribes(ctx)
	if err != nil {
		logger.WithContext(ctx).Error("[TrafficStatus] query active subscribes failed",
			logger.Field("error", err.Error()))
		return nil // 不返回 err,避免 asynq 重试雪崩
	}
	now := time.Now()
	cutOffHappened := false
	for _, sub := range subs {
		switch {
		case sub.CutOffAt != nil && now.After(*sub.CutOffAt) && !sub.Notified24h:
			l.markCutOff(ctx, sub)
			cutOffHappened = true
		case sub.ThrottledAt != nil && !sub.Notified12h && now.Sub(*sub.ThrottledAt) >= 12*time.Hour:
			l.mark12h(ctx, sub)
		case sub.ThrottledAt == nil && !sub.Notified90:
			l.maybeWarn90(ctx, sub)
		}
	}
	if cutOffHappened {
		// 节点端立即同步,被 cut off 的订阅设备从 user list 消失
		l.invalidateNodeCache(ctx)
	}
	return nil
}

// activeSubscribes — 取所有 status=1 且 expire_time > now() 的订阅。
// 限制 1000 条/批,够覆盖小到中型机场;大型场景需做分页(后续 Phase 优化)。
func (l *TrafficStatusLogic) activeSubscribes(ctx context.Context) ([]*user.Subscribe, error) {
	var list []*user.Subscribe
	err := l.svc.DB.WithContext(ctx).
		Model(&user.Subscribe{}).
		Where("`status` = ? AND `expire_time` > ?", 1, time.Now()).
		Order("id ASC").
		Limit(2000).
		Find(&list).Error
	return list, err
}

func (l *TrafficStatusLogic) maybeWarn90(ctx context.Context, sub *user.Subscribe) {
	quota := sub.Traffic + sub.TrafficAddon
	if quota <= 0 {
		return
	}
	used := sub.Download + sub.Upload
	if float64(used)*100 < float64(quota)*90 {
		return
	}
	sub.Notified90 = true
	if err := l.svc.UserModel.UpdateSubscribe(ctx, sub); err != nil {
		logger.WithContext(ctx).Error("[TrafficStatus] mark notified_90 failed",
			logger.Field("user_subscribe_id", sub.Id), logger.Field("error", err.Error()))
		return
	}
	// 通知由 Phase 6 hook 邮件 — 此处先 audit
	_ = l.svc.AuditModel.AppendDetail(ctx, &auditmodel.AuditLog{
		UserId: sub.UserId, Actor: auditmodel.ActorSystem,
		Action: "notify_traffic_90",
		Target: subscribeTarget(sub.Id),
	}, map[string]interface{}{"used": used, "quota": quota})
	enqueueNotice(l.svc, sub.UserId, sub.Id, "traffic_90")
}

func (l *TrafficStatusLogic) mark12h(ctx context.Context, sub *user.Subscribe) {
	sub.Notified12h = true
	if err := l.svc.UserModel.UpdateSubscribe(ctx, sub); err != nil {
		logger.WithContext(ctx).Error("[TrafficStatus] mark notified_12h failed",
			logger.Field("user_subscribe_id", sub.Id), logger.Field("error", err.Error()))
		return
	}
	_ = l.svc.AuditModel.Append(ctx, &auditmodel.AuditLog{
		UserId: sub.UserId, Actor: auditmodel.ActorSystem,
		Action: "notify_throttle_12h",
		Target: subscribeTarget(sub.Id),
	})
	enqueueNotice(l.svc, sub.UserId, sub.Id, "throttle_12h")
}

func (l *TrafficStatusLogic) markCutOff(ctx context.Context, sub *user.Subscribe) {
	sub.Notified24h = true
	if err := l.svc.UserModel.UpdateSubscribe(ctx, sub); err != nil {
		logger.WithContext(ctx).Error("[TrafficStatus] mark notified_24h failed",
			logger.Field("user_subscribe_id", sub.Id), logger.Field("error", err.Error()))
		return
	}
	_ = l.svc.AuditModel.Append(ctx, &auditmodel.AuditLog{
		UserId: sub.UserId, Actor: auditmodel.ActorSystem,
		Action: auditmodel.ActionThrottleCutOff,
		Target: subscribeTarget(sub.Id),
	})
	enqueueNotice(l.svc, sub.UserId, sub.Id, "cutoff")
}

func (l *TrafficStatusLogic) invalidateNodeCache(ctx context.Context) {
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

// subscribeTarget — audit_log.target 标准化前缀。
func subscribeTarget(id int64) string {
	return "user_subscribe:" + intToStr(id)
}

func intToStr(n int64) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{digits[n%10]}, buf...)
		n /= 10
	}
	if neg {
		return "-" + string(buf)
	}
	return string(buf)
}

// enqueueNotice — Phase 6 通知系统的胶水。先做轻量埋点,Phase 6 再实际发送。
// 写到一个独立 redis list,Phase 6 cron consumer 拉取。
func enqueueNotice(svcCtx *svc.ServiceContext, userId, userSubscribeId int64, templateKey string) {
	// 暂存到 Redis list,Phase 6 提供专门 worker。
	_ = svcCtx.Redis.LPush(context.Background(),
		"notice:queue",
		// 极简编码:user_id|user_subscribe_id|template_key
		intToStr(userId)+"|"+intToStr(userSubscribeId)+"|"+templateKey,
	).Err()
}
