package device

// V4.3 加购单设备(决策 4 / 8 / 10)
// - 校验:device_count + 1 ≤ subscribe.max_device_count
// - 计费:amount = round(unit_price_per_device × max(0, expire-now)/total_days)
// - 事务:扣余额 → user_subscribe.device_count++ → INSERT user_subscribe_device → DEL Redis user list cache
// - audit_log 异步追加

import (
	"context"
	"fmt"
	"math"
	"time"

	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	subscribepkg "github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/model/user"
	subscribelogic "github.com/perfect-panel/server/internal/logic/public/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type AddSubscribeDeviceLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddSubscribeDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddSubscribeDeviceLogic {
	return &AddSubscribeDeviceLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddSubscribeDeviceLogic) AddSubscribeDevice(req *types.AddSubscribeDeviceRequest) (*types.AddSubscribeDeviceResponse, error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}

	// 1) 取目标 user_subscribe + plan
	userSub, err := l.svcCtx.UserModel.FindOneSubscribe(l.ctx, req.UserSubscribeId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find user subscribe: %v", err)
	}
	if userSub.UserId != u.Id {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "subscribe not owned by user")
	}
	if userSub.Status != 1 {
		return nil, errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams, "订阅当前状态不允许加购设备"),
			"subscribe not active")
	}
	plan, err := l.svcCtx.SubscribeModel.FindOne(l.ctx, userSub.SubscribeId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe plan: %v", err)
	}
	if plan.UnitPricePerDevice <= 0 {
		return nil, errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams, "当前套餐不支持加购设备,请联系管理员配置每台设备单价"),
			"subscribe plan does not support device-billing")
	}
	// 数量校验:默认 1 台,允许批量加购,但不能超过套餐 MaxDeviceCount。
	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}
	if qty > 100 {
		return nil, errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams, "单次加购数量不能超过 100 台"),
			"quantity too large")
	}
	if plan.MaxDeviceCount > 0 && userSub.DeviceCount+qty > plan.MaxDeviceCount {
		return nil, errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams,
				fmt.Sprintf("已达套餐最大设备数 %d,无法继续加购 %d 台(当前 %d 台)",
					plan.MaxDeviceCount, qty, userSub.DeviceCount)),
			"device count exceeds max")
	}

	// 2) 比例计费(决策 4):金额 = 单台 × 数量
	now := time.Now()
	perDevice, ratio := proRatedDevicePrice(plan, userSub, now)
	amount := perDevice * qty
	if u.Balance < amount {
		return nil, errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams,
				fmt.Sprintf("余额不足,需 ¥%.2f,当前 ¥%.2f", float64(amount)/100, float64(u.Balance)/100)),
			"insufficient balance")
	}

	// 3) 事务:扣款 + device_count += qty + 批量新设备槽
	devices := make([]*user.SubscribeDevice, 0, qty)
	for i := int64(1); i <= qty; i++ {
		devices = append(devices, &user.SubscribeDevice{
			UserSubscribeId: userSub.Id,
			UserId:          u.Id,
			DeviceName:      fmt.Sprintf("设备 %d", userSub.DeviceCount+i),
			Status:          1,
			IsAddon:         true, // V4.3:用户主动加购,允许后续删除
		})
	}
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// a) 扣款
		u.Balance -= amount
		if e := l.svcCtx.UserModel.Update(l.ctx, u, tx); e != nil {
			return e
		}
		// b) device_count += qty
		userSub.DeviceCount += qty
		if e := l.svcCtx.UserModel.UpdateSubscribe(l.ctx, userSub, tx); e != nil {
			return e
		}
		// c) 批量新设备槽
		if e := l.svcCtx.UserModel.BatchInsertSubscribeDevices(l.ctx, devices, tx); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		l.Errorw("[AddSubscribeDevice] tx failed", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "add device tx: %v", err)
	}

	// 4) 失活 server user list 缓存
	_ = l.svcCtx.Redis.Del(l.ctx, subscribelogic.RedisServerUserListGlobalKey).Err()

	// V4.3 决策 21:邀请佣金,按总金额计算
	go subscribelogic.GrantCommission(l.svcCtx, u, plan, amount, "add_device")

	// 5) audit:一条日志记录批量加购
	deviceIds := make([]int64, 0, len(devices))
	deviceNames := make([]string, 0, len(devices))
	for _, d := range devices {
		deviceIds = append(deviceIds, d.Id)
		deviceNames = append(deviceNames, d.DeviceName)
	}
	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		UserId:  u.Id,
		Actor:   auditmodel.ActorUser,
		ActorId: u.Id,
		Action:  auditmodel.ActionAddDevice,
		Target:  fmt.Sprintf("user_subscribe:%d", userSub.Id),
	}, map[string]interface{}{
		"amount":       amount,
		"ratio_bp":     ratio,
		"quantity":     qty,
		"device_ids":   deviceIds,
		"device_names": deviceNames,
	})

	first := devices[0]
	urls := subscribelogic.BuildSubscribeURLs(l.svcCtx, first.Token)
	primary := ""
	if len(urls) > 0 {
		primary = urls[0]
	} else {
		primary = subscribelogic.BuildSubscribeURL(l.svcCtx, first.Token)
	}
	return &types.AddSubscribeDeviceResponse{
		DeviceId:      first.Id,
		Token:         first.Token,
		UUID:          first.UUID,
		Amount:        amount,
		Ratio:         ratio,
		SubscribeUrl:  primary,
		SubscribeUrls: urls,
		Quantity:      qty,
		DeviceIds:     deviceIds,
	}, nil
}

// proRatedDevicePrice — 4.2: amount = round(unit_price × max(0, expire-now)/total_days)。
// total_days 取 (expire - start) 的天数;不足 1 天按 1 天处理。返回值二元组:amount(分),ratio(万分比)。
func proRatedDevicePrice(plan *subscribepkg.Subscribe, sub *user.Subscribe, now time.Time) (int64, int64) {
	if sub.ExpireTime.IsZero() || !sub.ExpireTime.After(now) {
		return 0, 0
	}
	totalDays := sub.ExpireTime.Sub(sub.StartTime).Hours() / 24.0
	if totalDays < 1 {
		totalDays = 1
	}
	remainDays := sub.ExpireTime.Sub(now).Hours() / 24.0
	if remainDays < 0 {
		remainDays = 0
	}
	ratio := remainDays / totalDays
	if ratio > 1 {
		ratio = 1
	}
	amount := int64(math.Round(float64(plan.UnitPricePerDevice) * ratio))
	return amount, int64(math.Round(ratio * 10000))
}

// 用 tool 包做导入存在性占位(后续 Phase 会实际使用)。
var _ = tool.GenerateUUIDv4
