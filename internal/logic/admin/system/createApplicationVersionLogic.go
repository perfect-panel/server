package system

import (
	"context"

	"github.com/perfect-panel/ppanel-server/internal/model/application"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type CreateApplicationVersionLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create application version
func NewCreateApplicationVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateApplicationVersionLogic {
	return &CreateApplicationVersionLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateApplicationVersionLogic) CreateApplicationVersion(req *types.CreateApplicationVersionRequest) error {
	create := &application.ApplicationVersion{
		Url:           req.Url,
		Platform:      req.Platform,
		Version:       req.Version,
		Description:   req.Description,
		IsDefault:     req.IsDefault,
		ApplicationId: req.ApplicationId,
	}
	err := l.svcCtx.ApplicationModel.InsertVersion(l.ctx, create)
	if err != nil {
		l.Errorw("[CreateApplicationVersion] create application version error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create application version error: %v", err)
	}
	return nil
}
