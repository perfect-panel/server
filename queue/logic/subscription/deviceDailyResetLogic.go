package subscription

// V4.3 设备 today_traffic 凌晨归零 + reset_count 计数器轻量维护。
// 每天 00:00 跑一次。代价小:只 UPDATE,设备数远小于流量条数。

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

type DeviceDailyResetLogic struct {
	svc *svc.ServiceContext
}

func NewDeviceDailyResetLogic(svc *svc.ServiceContext) *DeviceDailyResetLogic {
	return &DeviceDailyResetLogic{svc: svc}
}

func (l *DeviceDailyResetLogic) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	res := l.svc.DB.WithContext(ctx).
		Model(&user.SubscribeDevice{}).
		Where("today_traffic > 0 OR reset_count_day > 0 OR reset_count_hour > 0").
		Updates(map[string]interface{}{
			"today_traffic":    0,
			"reset_count_day":  0,
			"reset_count_hour": 0,
		})
	if res.Error != nil {
		logger.WithContext(ctx).Error("[DeviceDailyReset] failed",
			logger.Field("error", res.Error.Error()))
		return nil
	}
	logger.WithContext(ctx).Info("[DeviceDailyReset] done",
		logger.Field("rows", res.RowsAffected))
	return nil
}
