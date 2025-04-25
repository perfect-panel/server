package system

import (
	"context"
	"strings"

	"github.com/perfect-panel/ppanel-server/internal/model/application"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdateApplicationConfigLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// update application config
func NewUpdateApplicationConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateApplicationConfigLogic {
	return &UpdateApplicationConfigLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateApplicationConfigLogic) UpdateApplicationConfig(req *types.ApplicationConfig) error {
	err := l.svcCtx.ApplicationModel.UpdateConfig(l.ctx, &application.ApplicationConfig{
		Id:                     1,
		AppId:                  req.AppId,
		EncryptionKey:          req.EncryptionKey,
		EncryptionMethod:       req.EncryptionMethod,
		Domains:                strings.Join(req.Domains, ";"),
		StartupPicture:         req.StartupPicture,
		StartupPictureSkipTime: req.StartupPictureSkipTime,
	})
	if err != nil {
		l.Errorw("[UpdateApplicationConfig] Database Error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update app config error: %v", err.Error())
	}
	return nil
}
