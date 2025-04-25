package order

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/ppanel-server/internal/model/order"
	"github.com/perfect-panel/ppanel-server/pkg/tool"

	"github.com/perfect-panel/ppanel-server/pkg/constant"

	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type PreCreateOrderLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Pre create order
func NewPreCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreCreateOrderLogic {
	return &PreCreateOrderLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreCreateOrderLogic) PreCreateOrder(req *types.PurchaseOrderRequest) (resp *types.PreOrderResponse, err error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	// find subscribe plan
	sub, err := l.svcCtx.SubscribeModel.FindOne(l.ctx, req.SubscribeId)
	if err != nil {
		l.Error("[PreCreateOrder] Database query error", logger.Field("error", err.Error()), logger.Field("subscribe_id", req.SubscribeId))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe error: %v", err.Error())
	}
	var discount float64 = 1
	if sub.Discount != "" {
		var dis []types.SubscribeDiscount
		_ = json.Unmarshal([]byte(sub.Discount), &dis)
		discount = getDiscount(dis, req.Quantity)
	}
	price := sub.UnitPrice * req.Quantity
	amount := int64(float64(price) * discount)
	discountAmount := price - amount
	var coupon int64
	if req.Coupon != "" {
		couponInfo, err := l.svcCtx.CouponModel.FindOneByCode(l.ctx, req.Coupon)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.Wrapf(xerr.NewErrCode(xerr.CouponNotExist), "coupon not found")
			}
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find coupon error: %v", err.Error())
		}
		if couponInfo.Count <= couponInfo.UsedCount {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.CouponInsufficientUsage), "coupon used")
		}
		couponSub := tool.StringToInt64Slice(couponInfo.Subscribe)
		if len(couponSub) > 0 && !tool.Contains(couponSub, req.SubscribeId) {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.CouponNotApplicable), "coupon not match")
		}
		var count int64
		err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			return tx.Model(&order.Order{}).Where("user_id = ? and coupon = ?", u.Id, req.Coupon).Count(&count).Error
		})

		if err != nil {
			l.Errorw("[Purchase] Database query error", logger.Field("error", err.Error()), logger.Field("user_id", u.Id), logger.Field("coupon", req.Coupon))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find coupon error: %v", err.Error())
		}
		if count >= couponInfo.UserLimit {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.CouponInsufficientUsage), "coupon limit exceeded")
		}
		coupon = calculateCoupon(amount, couponInfo)
	}
	amount -= coupon

	var deductionAmount int64
	// Check user deduction amount
	if u.GiftAmount > 0 {
		if u.GiftAmount >= amount {
			deductionAmount = amount
			amount = 0
		} else {
			deductionAmount = u.GiftAmount
			amount -= u.GiftAmount
		}
	}

	payment, err := l.svcCtx.PaymentModel.FindOne(l.ctx, req.Payment)
	if err != nil {
		l.Logger.Error("[PreCreateOrder] Database query error", logger.Field("error", err.Error()), logger.Field("payment", req.Payment))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find payment method error: %v", err.Error())
	}
	var feeAmount int64
	// Calculate the handling fee
	if amount > 0 {
		feeAmount = calculateFee(amount, payment)
	}
	amount += feeAmount

	resp = &types.PreOrderResponse{
		Price:          price,
		Amount:         amount,
		Discount:       discountAmount,
		GiftAmount:     deductionAmount,
		Coupon:         req.Coupon,
		CouponDiscount: coupon,
		FeeAmount:      feeAmount,
	}
	return
}
