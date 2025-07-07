package user

import (
	"context"

	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/perfect-panel/server/internal/model/user"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type UnsubscribeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUnsubscribeLogic creates a new instance of UnsubscribeLogic for handling subscription cancellation
func NewUnsubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnsubscribeLogic {
	return &UnsubscribeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Unsubscribe handles the subscription cancellation process with proper refund distribution
// It prioritizes refunding to gift amount for balance-paid orders, then to regular balance
func (l *UnsubscribeLogic) Unsubscribe(req *types.UnsubscribeRequest) error {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	// Calculate the remaining amount to refund based on unused subscription time/traffic
	remainingAmount, err := CalculateRemainingAmount(l.ctx, l.svcCtx, req.Id)
	if err != nil {
		return err
	}

	// Process unsubscription in a database transaction to ensure data consistency
	err = l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		// Find and update subscription status to cancelled (status = 4)
		var userSub user.Subscribe
		if err = db.Model(&user.Subscribe{}).Where("id = ?", req.Id).First(&userSub).Error; err != nil {
			return err
		}
		userSub.Status = 4 // Set status to cancelled
		if err = l.svcCtx.UserModel.UpdateSubscribe(l.ctx, &userSub); err != nil {
			return err
		}

		// Query the original order information to determine refund strategy
		orderInfo, err := l.svcCtx.OrderModel.FindOne(l.ctx, userSub.OrderId)
		if err != nil {
			return err
		}
		// Calculate refund distribution based on payment method and gift amount priority
		var balance, gift int64
		if orderInfo.Method == "balance" {
			// For balance-paid orders, prioritize refunding to gift amount first
			if orderInfo.GiftAmount >= remainingAmount {
				// Gift amount covers the entire refund - refund all to gift balance
				gift = remainingAmount
				balance = u.Balance // Regular balance remains unchanged
			} else {
				// Gift amount insufficient - refund to gift first, remainder to regular balance
				gift = orderInfo.GiftAmount
				balance = u.Balance + (remainingAmount - orderInfo.GiftAmount)
			}
		} else {
			// For non-balance payment orders, refund entirely to regular balance
			balance = remainingAmount + u.Balance
			gift = 0
		}

		// Create balance log entry only if there's an actual regular balance refund
		balanceRefundAmount := balance - u.Balance
		if balanceRefundAmount > 0 {
			balanceLog := user.BalanceLog{
				UserId:  userSub.UserId,
				OrderId: userSub.OrderId,
				Amount:  balanceRefundAmount,
				Type:    4, // Type 4 represents refund transaction
				Balance: balance,
			}
			if err := db.Model(&user.BalanceLog{}).Create(&balanceLog).Error; err != nil {
				return err
			}
		}

		// Create gift amount log entry if there's a gift balance refund
		if gift > 0 {
			giftLog := user.GiftAmountLog{
				UserId:          userSub.UserId,
				UserSubscribeId: userSub.Id,
				OrderNo:         orderInfo.OrderNo,
				Type:            1, // Type 1 represents gift amount increase
				Amount:          gift,
				Balance:         u.GiftAmount + gift,
				Remark:          "Unsubscribe refund",
			}
			if err := db.Model(&user.GiftAmountLog{}).Create(&giftLog).Error; err != nil {
				return err
			}
			// Update user's gift amount
			u.GiftAmount += gift
		}

		// Update user's regular balance and save changes to database
		u.Balance = balance
		return l.svcCtx.UserModel.Update(l.ctx, u)
	})

	return err
}
