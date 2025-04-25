package system

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type DeleteApplicationLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteApplicationLogic {
	return &DeleteApplicationLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteApplicationLogic) DeleteApplication(req *types.DeleteApplicationRequest) error {
	// delete application
	err := l.svcCtx.ApplicationModel.Delete(l.ctx, req.Id)
	if err != nil {
		l.Errorw("[DeleteApplicationLogic] delete application error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete application error: %v", err.Error())
	}
	return nil
}
