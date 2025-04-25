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

// NewUnsubscribeLogic Unsubscribe
func NewUnsubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnsubscribeLogic {
	return &UnsubscribeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnsubscribeLogic) Unsubscribe(req *types.UnsubscribeRequest) error {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	remainingAmount, err := CalculateRemainingAmount(l.ctx, l.svcCtx, req.Id)
	if err != nil {
		return err
	}
	// update user subscribe
	err = l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		var userSub user.Subscribe
		if err := db.Model(&user.Subscribe{}).Where("id = ?", req.Id).First(&userSub).Error; err != nil {
			return err
		}
		userSub.Status = 4
		if err := l.svcCtx.UserModel.UpdateSubscribe(l.ctx, &userSub); err != nil {
			return err
		}
		balance := remainingAmount + u.Balance
		// insert deduction log
		balanceLog := user.BalanceLog{
			UserId:  userSub.UserId,
			OrderId: userSub.OrderId,
			Amount:  remainingAmount,
			Type:    4,
			Balance: balance,
		}
		if err := db.Model(&user.BalanceLog{}).Create(&balanceLog).Error; err != nil {
			return err
		}
		// update user balance
		u.Balance = balance
		return l.svcCtx.UserModel.Update(l.ctx, u)
	})

	return err
}
