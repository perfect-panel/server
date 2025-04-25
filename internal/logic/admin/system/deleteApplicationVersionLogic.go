package system

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type DeleteApplicationVersionLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete application
func NewDeleteApplicationVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteApplicationVersionLogic {
	return &DeleteApplicationVersionLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteApplicationVersionLogic) DeleteApplicationVersion(req *types.DeleteApplicationVersionRequest) error {
	// delete application
	err := l.svcCtx.ApplicationModel.DeleteVersion(l.ctx, req.Id)
	if err != nil {
		l.Errorw("[DeleteApplicationVersion] delete application version error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete application version error: %v", err.Error())
	}
	return nil
}
