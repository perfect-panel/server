package redemption

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ToggleRedemptionCodeStatusLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Toggle redemption code status
func NewToggleRedemptionCodeStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ToggleRedemptionCodeStatusLogic {
	return &ToggleRedemptionCodeStatusLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ToggleRedemptionCodeStatusLogic) ToggleRedemptionCodeStatus(req *types.ToggleRedemptionCodeStatusRequest) error {
	// Find redemption code
	codeInfo, err := l.svcCtx.RedemptionCodeModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorw("[ToggleRedemptionCodeStatus] Redemption code not found", logger.Field("id", req.Id))
			return errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "redemption code not found")
		}
		l.Errorw("[ToggleRedemptionCodeStatus] Database Error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find redemption code error: %v", err.Error())
	}

	// Update status
	codeInfo.Status = req.Status

	err = l.svcCtx.RedemptionCodeModel.Update(l.ctx, codeInfo)
	if err != nil {
		l.Errorw("[ToggleRedemptionCodeStatus] Database Error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update redemption code status error: %v", err.Error())
	}

	l.Infow("[ToggleRedemptionCodeStatus] Successfully toggled redemption code status",
		logger.Field("id", req.Id),
		logger.Field("status", req.Status))

	return nil
}
