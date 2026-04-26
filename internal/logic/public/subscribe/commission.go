package subscribe

// V4.3 决策 21:邀请佣金 10%(含加购)。
//
// 优先级:plan.commission_rate(per-package 覆写)→ user.referer.referral_percentage
// → 全局 config.Invite.ReferralPercentage。
//
// 与 activateOrderLogic.handleCommission 区别:
// - 后者通过订单流(order.amount, order.fee_amount)计算
// - 加购流量包/加购设备没走 order 表(直接 balance 扣款),所以这里用纯 helper
//   重写一遍。amount 入参就是用户实付金额(分)。

import (
	"context"
	"encoding/json"

	logmodel "github.com/perfect-panel/server/internal/model/log"
	subscribepkg "github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"gorm.io/gorm"
)

// GrantCommission 给邀请人分佣。
// - 不阻塞主流程,异步运行。
// - 任何失败只记录日志,绝不回滚 user balance/订阅状态。
//
// 导出符号给 internal/logic/public/{device,trafficaddon} 复用(加购单设备 / 流量包亦计佣金)。
func GrantCommission(svcCtx *svc.ServiceContext, buyer *user.User, plan *subscribepkg.Subscribe, amount int64, source string) {
	if buyer == nil || buyer.RefererId == 0 || amount <= 0 {
		return
	}
	ctx := context.Background()
	referer, err := svcCtx.UserModel.FindOne(ctx, buyer.RefererId)
	if err != nil || referer == nil {
		return
	}
	rate := pickCommissionRate(plan, referer, svcCtx)
	if rate <= 0 {
		return
	}
	commission := amount * rate / 100
	if commission <= 0 {
		return
	}
	err = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		referer.Commission += commission
		if e := svcCtx.UserModel.Update(ctx, referer, tx); e != nil {
			return e
		}
		// 写一条 commission systemlog,沿用现有 Type=33 / TypeCommission 编码
		entry := &logmodel.Commission{
			Type:      logmodel.CommissionTypePurchase, // 加购也归到购买类型
			Amount:    commission,
			OrderNo:   source, // 这里没有 OrderNo,占位 source
			Timestamp: 0,
		}
		content, _ := json.Marshal(entry)
		return tx.Create(&logmodel.SystemLog{
			Type:     logmodel.TypeCommission.Uint8(),
			ObjectID: referer.Id,
			Content:  string(content),
		}).Error
	})
	if err != nil {
		logger.WithContext(ctx).Error("[GrantCommission] tx failed",
			logger.Field("error", err.Error()),
			logger.Field("buyer_id", buyer.Id),
			logger.Field("referer_id", referer.Id),
			logger.Field("source", source))
		return
	}
	_ = svcCtx.UserModel.UpdateUserCache(ctx, referer)
}

// pickCommissionRate — 三级 fallback 取百分比。
func pickCommissionRate(plan *subscribepkg.Subscribe, referer *user.User, svcCtx *svc.ServiceContext) int64 {
	if plan != nil && plan.CommissionRate > 0 {
		return plan.CommissionRate
	}
	if referer.ReferralPercentage > 0 {
		return int64(referer.ReferralPercentage)
	}
	return int64(svcCtx.Config.Invite.ReferralPercentage)
}
