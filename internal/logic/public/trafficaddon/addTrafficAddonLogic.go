package trafficaddon

// V4.3 流量加购(决策 7 / 16 / 20 / 40)
// - addon_bytes 必须为 unit_size 的整数倍
// - amount = (bytes / unit_size) × unit_price
// - 事务:扣余额 → user_subscribe.traffic_addon += bytes →
//          throttled_at/cut_off_at = NULL → notified_100/12h/24h = 0
//          → INSERT traffic_addon_order
// - 失活 server user list 缓存 → 节点 next pull 即解除限速

import (
	"context"
	"fmt"

	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	subscribelogic "github.com/perfect-panel/server/internal/logic/public/subscribe"
	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type AddTrafficAddonLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddTrafficAddonLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddTrafficAddonLogic {
	return &AddTrafficAddonLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddTrafficAddonLogic) AddTrafficAddon(req *types.AddTrafficAddonRequest) (*types.AddTrafficAddonResponse, error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	if req.AddonBytes <= 0 {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "addon_bytes must be > 0")
	}

	userSub, err := l.svcCtx.UserModel.FindOneSubscribe(l.ctx, req.UserSubscribeId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe: %v", err)
	}
	if userSub.UserId != u.Id {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "subscribe not owned by user")
	}

	plan, err := l.svcCtx.SubscribeModel.FindOne(l.ctx, userSub.SubscribeId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find plan: %v", err)
	}
	if plan.TrafficAddonUnitPrice <= 0 || plan.TrafficAddonUnitSize <= 0 {
		return nil, errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams, "当前套餐不支持加购流量,请联系管理员配置流量包单价"),
			"subscribe plan does not support traffic addon")
	}
	if req.AddonBytes%plan.TrafficAddonUnitSize != 0 {
		return nil, errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams,
				fmt.Sprintf("加购流量必须是 %d 字节的整数倍", plan.TrafficAddonUnitSize)),
			"addon_bytes not multiple of unit_size")
	}

	units := req.AddonBytes / plan.TrafficAddonUnitSize
	amount := plan.TrafficAddonUnitPrice * units
	if u.Balance < amount {
		return nil, errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams,
				fmt.Sprintf("余额不足,需 ¥%.2f,当前 ¥%.2f", float64(amount)/100, float64(u.Balance)/100)),
			"insufficient balance")
	}

	addonOrder := &order.TrafficAddonOrder{
		UserId:          u.Id,
		UserSubscribeId: userSub.Id,
		AddonBytes:      req.AddonBytes,
		Amount:          amount,
		UnitPrice:       plan.TrafficAddonUnitPrice,
		UnitSize:        plan.TrafficAddonUnitSize,
	}
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// a) 扣款
		u.Balance -= amount
		if e := l.svcCtx.UserModel.Update(l.ctx, u, tx); e != nil {
			return e
		}
		// b) 流量池 + 重置限速/通知标志
		userSub.TrafficAddon += req.AddonBytes
		userSub.ThrottledAt = nil
		userSub.CutOffAt = nil
		userSub.Notified100 = false
		userSub.Notified12h = false
		userSub.Notified24h = false
		// 90% 阈值是否重置取决于加购后是否回到 90% 以下
		if !overUsedThreshold(userSub, plan, 0.90) {
			userSub.Notified90 = false
		}
		if e := l.svcCtx.UserModel.UpdateSubscribe(l.ctx, userSub, tx); e != nil {
			return e
		}
		// c) 流水
		if e := l.svcCtx.OrderModel.InsertTrafficAddonOrder(l.ctx, addonOrder, tx); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		l.Errorw("[AddTrafficAddon] tx failed", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "addon tx: %v", err)
	}

	// 节点端立即解除限速
	_ = l.svcCtx.Redis.Del(l.ctx, subscribelogic.RedisServerUserListGlobalKey).Err()

	// V4.3 决策 21:加购流量包同样计佣金(含加购)。
	go subscribelogic.GrantCommission(l.svcCtx, u, plan, amount, "addon_traffic")

	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		UserId:  u.Id,
		Actor:   auditmodel.ActorUser,
		ActorId: u.Id,
		Action:  auditmodel.ActionAddonTraffic,
		Target:  fmt.Sprintf("user_subscribe:%d", userSub.Id),
	}, map[string]interface{}{
		"addon_bytes":     req.AddonBytes,
		"amount":          amount,
		"addon_order_id":  addonOrder.Id,
		"throttled_reset": true,
	})

	return &types.AddTrafficAddonResponse{
		AddonOrderId: addonOrder.Id,
		AddonBytes:   req.AddonBytes,
		Amount:       amount,
		TrafficTotal: userSub.Traffic + userSub.TrafficAddon,
	}, nil
}

// overUsedThreshold returns true when (download+upload) >= threshold * (traffic + addon).
func overUsedThreshold(sub *user.Subscribe, _ interface{}, threshold float64) bool {
	quota := sub.Traffic + sub.TrafficAddon
	if quota <= 0 {
		return false
	}
	used := sub.Download + sub.Upload
	return float64(used) >= threshold*float64(quota)
}
