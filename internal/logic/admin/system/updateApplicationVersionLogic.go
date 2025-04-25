package system

import (
	"context"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdateApplicationVersionLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update application version
func NewUpdateApplicationVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateApplicationVersionLogic {
	return &UpdateApplicationVersionLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateApplicationVersionLogic) UpdateApplicationVersion(req *types.UpdateApplicationVersionRequest) error {
	// find application
	app, err := l.svcCtx.ApplicationModel.FindOneVersion(l.ctx, req.Id)
	if err != nil {
		l.Errorw("[UpdateApplicationVersion] find application version error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find application error: %v", err.Error())
	}
	app.Url = req.Url
	app.Version = req.Version
	app.Description = req.Description
	app.IsDefault = req.IsDefault
	err = l.svcCtx.ApplicationModel.UpdateVersion(l.ctx, app)
	if err != nil {
		l.Errorw("[UpdateApplicationVersion] update application version error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update application version error: %v", err.Error())
	}
	return nil
}
