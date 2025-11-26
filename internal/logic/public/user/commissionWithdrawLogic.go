package user

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
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
	// todo: add your logic here and delete this line

	return
}
