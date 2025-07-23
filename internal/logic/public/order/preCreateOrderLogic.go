package order

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/pkg/tool"

	"github.com/perfect-panel/server/pkg/constant"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type PreCreateOrderLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPreCreateOrderLogic creates a new pre-create order logic instance for order preview operations.
// It initializes the logger with context and sets up the service context for database operations.
func NewPreCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreCreateOrderLogic {
	return &PreCreateOrderLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PreCreateOrder calculates order pricing preview including discounts, coupons, gift amounts, and fees
// without actually creating an order. It validates subscription plans, coupons, and payment methods
// to provide accurate pricing information for the frontend order preview.
func (l *PreCreateOrderLogic) PreCreateOrder(req *types.PurchaseOrderRequest) (resp *types.PreOrderResponse, err error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}

	if req.Quantity <= 0 {
		l.Debugf("[PreCreateOrder] Quantity is less than or equal to 0, setting to 1")
		req.Quantity = 1
	}

	// find subscribe plan
	sub, err := l.svcCtx.SubscribeModel.FindOne(l.ctx, req.SubscribeId)
	if err != nil {
		l.Errorw("[PreCreateOrder] Database query error", logger.Field("error", err.Error()), logger.Field("subscribe_id", req.SubscribeId))
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
	var couponAmount int64
	if req.Coupon != "" {
		couponInfo, err := l.svcCtx.CouponModel.FindOneByCode(l.ctx, req.Coupon)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.Wrapf(xerr.NewErrCode(xerr.CouponNotExist), "coupon not found")
			}
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find coupon error: %v", err.Error())
		}
		if couponInfo.Count > 0 && couponInfo.Count <= couponInfo.UsedCount {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.CouponAlreadyUsed), "coupon used")
		}
		var count int64
		err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			return tx.Model(&order.Order{}).Where("user_id = ? and coupon = ?", u.Id, req.Coupon).Count(&count).Error
		})

		if err != nil {
			l.Errorw("[PreCreateOrder] Database query error", logger.Field("error", err.Error()), logger.Field("user_id", u.Id), logger.Field("coupon", req.Coupon))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find coupon error: %v", err.Error())
		}

		if couponInfo.UserLimit > 0 && count >= couponInfo.UserLimit {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.CouponInsufficientUsage), "coupon limit exceeded")
		}

		couponSub := tool.StringToInt64Slice(couponInfo.Subscribe)
		if len(couponSub) > 0 && !tool.Contains(couponSub, req.SubscribeId) {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.CouponNotApplicable), "coupon not match")
		}
		couponAmount = calculateCoupon(amount, couponInfo)
	}
	amount -= couponAmount

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
	var feeAmount int64
	if req.Payment != 0 {
		payment, err := l.svcCtx.PaymentModel.FindOne(l.ctx, req.Payment)
		if err != nil {
			l.Errorw("[PreCreateOrder] Database query error", logger.Field("error", err.Error()), logger.Field("payment", req.Payment))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find payment method error: %v", err.Error())
		}
		// Calculate the handling fee
		if amount > 0 {
			feeAmount = calculateFee(amount, payment)
		}
		amount += feeAmount
	}

	resp = &types.PreOrderResponse{
		Price:          price,
		Amount:         amount,
		Discount:       discountAmount,
		GiftAmount:     deductionAmount,
		Coupon:         req.Coupon,
		CouponDiscount: couponAmount,
		FeeAmount:      feeAmount,
	}
	return
}
