package user

import (
	"context"
	"time"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type CommissionWithdrawLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Commission Withdraw
func NewCommissionWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommissionWithdrawLogic {
	return &CommissionWithdrawLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommissionWithdrawLogic) CommissionWithdraw(req *types.CommissionWithdrawRequest) (resp *types.WithdrawalLog, err error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}

	if u.Commission < req.Amount {
		logger.Errorf("User %d has insufficient commission balance: %.2f, requested: %.2f", u.Id, float64(u.Commission)/100, float64(req.Amount)/100)
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.UserCommissionNotEnough), "User %d has insufficient commission balance", u.Id)
	}

	tx := l.svcCtx.DB.WithContext(l.ctx).Begin()

	// update user commission balance
	u.Commission -= req.Amount
	if err = l.svcCtx.UserModel.Update(l.ctx, u, tx); err != nil {
		tx.Rollback()
		l.Errorf("Failed to update user %d commission balance: %v", u.Id, err)
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "Failed to update user %d commission balance: %v", u.Id, err)
	}

	// create withdrawal log
	logInfo := log.Commission{
		Type:      log.CommissionTypeConvertBalance,
		Amount:    req.Amount,
		Timestamp: time.Now().UnixMilli(),
	}
	b, err := logInfo.Marshal()

	if err != nil {
		tx.Rollback()
		l.Errorf("Failed to marshal commission log for user %d: %v", u.Id, err)
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Failed to marshal commission log for user %d: %v", u.Id, err)
	}

	err = tx.Model(log.SystemLog{}).Create(&log.SystemLog{
		Type:      log.TypeCommission.Uint8(),
		Date:      time.Now().Format("2006-01-02"),
		ObjectID:  u.Id,
		Content:   string(b),
		CreatedAt: time.Now(),
	}).Error

	if err != nil {
		tx.Rollback()
		l.Errorf("Failed to create commission log for user %d: %v", u.Id, err)
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "Failed to create commission log for user %d: %v", u.Id, err)
	}

	err = tx.Model(&user.Withdrawal{}).Create(&user.Withdrawal{
		UserId:  u.Id,
		Amount:  req.Amount,
		Content: req.Content,
		Status:  0,
		Reason:  "",
	}).Error

	if err != nil {
		tx.Rollback()
		l.Errorf("Failed to create withdrawal log for user %d: %v", u.Id, err)
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "Failed to create withdrawal log for user %d: %v", u.Id, err)
	}
	if err = tx.Commit().Error; err != nil {
		l.Errorf("Transaction commit failed for user %d withdrawal: %v", u.Id, err)
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Transaction commit failed for user %d withdrawal: %v", u.Id, err)
	}

	return &types.WithdrawalLog{
		UserId:    u.Id,
		Amount:    req.Amount,
		Content:   req.Content,
		Status:    0,
		Reason:    "",
		CreatedAt: time.Now().UnixMilli(),
	}, nil
}
