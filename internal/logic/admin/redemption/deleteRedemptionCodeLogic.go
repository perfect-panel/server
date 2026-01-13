package redemption

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type DeleteRedemptionCodeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete redemption code
func NewDeleteRedemptionCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRedemptionCodeLogic {
	return &DeleteRedemptionCodeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRedemptionCodeLogic) DeleteRedemptionCode(req *types.DeleteRedemptionCodeRequest) error {
	err := l.svcCtx.RedemptionCodeModel.Delete(l.ctx, req.Id)
	if err != nil {
		l.Errorw("[DeleteRedemptionCode] Database Error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete redemption code error: %v", err.Error())
	}

	return nil
}
