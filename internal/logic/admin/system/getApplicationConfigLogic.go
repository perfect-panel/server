package system

import (
	"context"
	"strings"

	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type GetApplicationConfigLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// get application config
func NewGetApplicationConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApplicationConfigLogic {
	return &GetApplicationConfigLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetApplicationConfigLogic) GetApplicationConfig() (resp *types.ApplicationConfig, err error) {
	resp = &types.ApplicationConfig{}
	appConfig, err := l.svcCtx.ApplicationModel.FindOneConfig(l.ctx, 1)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		l.Errorw("[GetApplicationConfig] Database Error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get app config  error: %v", err.Error())
	}
	resp.AppId = appConfig.AppId
	resp.EncryptionKey = appConfig.EncryptionKey
	resp.EncryptionMethod = appConfig.EncryptionMethod
	resp.Domains = strings.Split(appConfig.Domains, ";")
	resp.StartupPicture = appConfig.StartupPicture
	resp.StartupPictureSkipTime = appConfig.StartupPictureSkipTime
	return
}
