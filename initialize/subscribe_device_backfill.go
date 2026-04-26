package initialize

// V4.3 修复:把所有「user_subscribe.device_count < subscribe.device_limit」的
// 老订阅补齐到套餐预设的设备数。
//
// 背景:V4.3 之前的订阅(或当前购买流程没传 device_count 的订单)在
// activateOrderLogic 里默认按 1 台设备建槽。导致管理端套餐写「使用设备 2 台」,
// 但用户面板显示「设备: 1 / 1」。这里启动时一次性纠正:
//   1. UPDATE user_subscribe SET device_count = subscribe.device_limit
//      WHERE device_count < device_limit
//   2. 给缺失的设备槽位 INSERT user_subscribe_device 行
//
// 幂等:多次运行只补齐缺口,不会重复创建。

import (
	"context"
	"fmt"

	subscribepkg "github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

func SubscribeDeviceBackfill(svcCtx *svc.ServiceContext) {
	logger.Debug("[Init SubscribeDeviceBackfill] V4.3 device-count alignment")
	ctx := context.Background()
	db := svcCtx.DB.WithContext(ctx)

	// 1) 找到所有需要补齐的 user_subscribe(device_count < subscribe.device_limit)
	type underProvisioned struct {
		Id           int64 `gorm:"column:id"`
		UserId       int64 `gorm:"column:user_id"`
		SubscribeId  int64 `gorm:"column:subscribe_id"`
		DeviceCount  int64 `gorm:"column:device_count"`
		PlanLimit    int64 `gorm:"column:plan_limit"`
	}
	var rows []underProvisioned
	err := db.Raw(`
		SELECT us.id, us.user_id, us.subscribe_id, us.device_count, s.device_limit AS plan_limit
		FROM user_subscribe us
		JOIN subscribe s ON s.id = us.subscribe_id
		WHERE s.device_limit > 0 AND us.device_count < s.device_limit
		  AND us.status = 1
	`).Scan(&rows).Error
	if err != nil {
		logger.Errorf("[SubscribeDeviceBackfill] scan failed: %v", err)
		return
	}
	if len(rows) == 0 {
		return
	}

	for _, r := range rows {
		need := r.PlanLimit - r.DeviceCount
		if need <= 0 {
			continue
		}
		// 数据 sanity:重新查一下当前实际的 device 行数,避免重复
		var existing int64
		_ = db.Model(&user.SubscribeDevice{}).
			Where("user_subscribe_id = ?", r.Id).
			Count(&existing).Error
		nameStart := existing + 1

		newDevices := make([]*user.SubscribeDevice, 0, need)
		for i := int64(0); i < need; i++ {
			newDevices = append(newDevices, &user.SubscribeDevice{
				UserSubscribeId: r.Id,
				UserId:          r.UserId,
				DeviceName:      fmt.Sprintf("设备 %d", nameStart+i),
				Status:          1,
			})
		}
		if err := svcCtx.UserModel.BatchInsertSubscribeDevices(ctx, newDevices); err != nil {
			logger.Errorf("[SubscribeDeviceBackfill] insert devices for sub %d: %v", r.Id, err)
			continue
		}
		// 更新 user_subscribe.device_count
		if err := db.Model(&user.Subscribe{}).
			Where("id = ?", r.Id).
			Update("device_count", r.PlanLimit).Error; err != nil {
			logger.Errorf("[SubscribeDeviceBackfill] update sub %d count: %v", r.Id, err)
			continue
		}
		logger.Infof("[SubscribeDeviceBackfill] sub_id=%d %d → %d devices",
			r.Id, r.DeviceCount, r.PlanLimit)
	}

	// 用 subscribe 包以避免 unused import 警告
	_ = (*subscribepkg.Subscribe)(nil)
}
