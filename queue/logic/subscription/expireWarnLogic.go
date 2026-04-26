package subscription

// V4.3 到期提醒 cron(决策 7.1 通知矩阵):每天 00:00 跑一次。
//
//   [active] ──── expire_time - now ≤ 3d ────▶ enqueue notice "expire_3d"
//                                              + Redis SETNX 保证只发 1 次
//             ──── expire_time - now ≤ 1d ────▶ enqueue notice "expire_1d"
//                                              + Redis SETNX
//
// 幂等用 Redis key `subscribe:expire_notified:<user_subscribe_id>:<key>`,TTL 略大于窗口。
// 不写到 user_subscribe 列(避免和限速 notified_* 字段语义混杂)。

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

const (
	expireWarn3dWindow = 3 * 24 * time.Hour
	expireWarn1dWindow = 1 * 24 * time.Hour
	// 幂等 TTL 取窗口的 1.5 倍,避免边界场景重复发送。
	expireWarn3dTTL = 4 * 24 * time.Hour
	expireWarn1dTTL = 2 * 24 * time.Hour
)

type ExpireWarnLogic struct {
	svc *svc.ServiceContext
}

func NewExpireWarnLogic(svc *svc.ServiceContext) *ExpireWarnLogic {
	return &ExpireWarnLogic{svc: svc}
}

func (l *ExpireWarnLogic) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	subs, err := l.candidates(ctx)
	if err != nil {
		logger.WithContext(ctx).Error("[ExpireWarn] query candidates failed",
			logger.Field("error", err.Error()))
		return nil
	}
	now := time.Now()
	for _, sub := range subs {
		remain := sub.ExpireTime.Sub(now)
		if remain <= 0 {
			continue // 已经过期由 CheckSubscription 处理
		}
		switch {
		case remain <= expireWarn1dWindow:
			l.tryEnqueue(ctx, sub, "expire_1d", expireWarn1dTTL)
		case remain <= expireWarn3dWindow:
			l.tryEnqueue(ctx, sub, "expire_3d", expireWarn3dTTL)
		}
	}
	return nil
}

// candidates — 取所有 status=1 且 expire_time 在 now ~ now+3d 区间的订阅。
// 单批限制 5000,日跑一次足够覆盖中型机场。
func (l *ExpireWarnLogic) candidates(ctx context.Context) ([]*user.Subscribe, error) {
	now := time.Now()
	upper := now.Add(expireWarn3dWindow)
	var list []*user.Subscribe
	err := l.svc.DB.WithContext(ctx).
		Model(&user.Subscribe{}).
		Where("`status` = ? AND `expire_time` > ? AND `expire_time` <= ?", 1, now, upper).
		Order("expire_time ASC").
		Limit(5000).
		Find(&list).Error
	return list, err
}

// tryEnqueue — Redis SETNX 保证同一订阅在 TTL 内只发 1 次该模板。
func (l *ExpireWarnLogic) tryEnqueue(ctx context.Context, sub *user.Subscribe, tplKey string, ttl time.Duration) {
	flagKey := fmt.Sprintf("subscribe:expire_notified:%d:%s", sub.Id, tplKey)
	ok, err := l.svc.Redis.SetNX(ctx, flagKey, 1, ttl).Result()
	if err != nil {
		logger.WithContext(ctx).Error("[ExpireWarn] setnx failed",
			logger.Field("user_subscribe_id", sub.Id),
			logger.Field("template", tplKey),
			logger.Field("error", err.Error()))
		return
	}
	if !ok {
		return // 已经发送过,跳过
	}
	enqueueNotice(l.svc, sub.UserId, sub.Id, tplKey)
}
