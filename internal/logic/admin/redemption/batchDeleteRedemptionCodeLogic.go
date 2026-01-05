package redemption

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type BatchDeleteRedemptionCodeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Batch delete redemption code
func NewBatchDeleteRedemptionCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchDeleteRedemptionCodeLogic {
	return &BatchDeleteRedemptionCodeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchDeleteRedemptionCodeLogic) BatchDeleteRedemptionCode(req *types.BatchDeleteRedemptionCodeRequest) error {
	err := l.svcCtx.RedemptionCodeModel.BatchDelete(l.ctx, req.Ids)
	if err != nil {
		l.Errorw("[BatchDeleteRedemptionCode] Database Error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "batch delete redemption code error: %v", err.Error())
	}

	return nil
}
