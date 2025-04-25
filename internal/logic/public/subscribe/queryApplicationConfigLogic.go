package subscribe

import (
	"context"

	"github.com/perfect-panel/ppanel-server/internal/model/application"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
)

type QueryApplicationConfigLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get application config
func NewQueryApplicationConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryApplicationConfigLogic {
	return &QueryApplicationConfigLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryApplicationConfigLogic) QueryApplicationConfig() (resp *types.ApplicationResponse, err error) {
	resp = &types.ApplicationResponse{}
	var applications []*application.Application
	err = l.svcCtx.ApplicationModel.Transaction(l.ctx, func(tx *gorm.DB) (err error) {
		return tx.Model(applications).Preload("ApplicationVersions").Find(&applications).Error
	})
	if err != nil {
		l.Errorw("[QueryApplicationConfig] get application error: ", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get application error: %v", err.Error())
	}

	if len(applications) == 0 {
		return resp, nil
	}

	for _, app := range applications {
		applicationResponse := types.ApplicationResponseInfo{
			Id:            app.Id,
			Name:          app.Name,
			Icon:          app.Icon,
			Description:   app.Description,
			SubscribeType: app.SubscribeType,
		}
		applicationVersions := app.ApplicationVersions
		if len(applicationVersions) != 0 {
			for _, applicationVersion := range applicationVersions {
				/*if !applicationVersion.IsDefault {
					continue
				}*/
				switch applicationVersion.Platform {
				case "ios":
					applicationResponse.Platform.IOS = append(applicationResponse.Platform.IOS, &types.ApplicationVersion{
						Id:          applicationVersion.Id,
						Url:         applicationVersion.Url,
						Version:     applicationVersion.Version,
						IsDefault:   applicationVersion.IsDefault,
						Description: applicationVersion.Description,
					})
				case "macos":
					applicationResponse.Platform.MacOS = append(applicationResponse.Platform.MacOS, &types.ApplicationVersion{
						Id:          applicationVersion.Id,
						Url:         applicationVersion.Url,
						Version:     applicationVersion.Version,
						IsDefault:   applicationVersion.IsDefault,
						Description: applicationVersion.Description,
					})
				case "linux":
					applicationResponse.Platform.Linux = append(applicationResponse.Platform.Linux, &types.ApplicationVersion{
						Id:          applicationVersion.Id,
						Url:         applicationVersion.Url,
						Version:     applicationVersion.Version,
						IsDefault:   applicationVersion.IsDefault,
						Description: applicationVersion.Description,
					})
				case "android":
					applicationResponse.Platform.Android = append(applicationResponse.Platform.Android, &types.ApplicationVersion{
						Id:          applicationVersion.Id,
						Url:         applicationVersion.Url,
						Version:     applicationVersion.Version,
						IsDefault:   applicationVersion.IsDefault,
						Description: applicationVersion.Description,
					})
				case "windows":
					applicationResponse.Platform.Windows = append(applicationResponse.Platform.Windows, &types.ApplicationVersion{
						Id:          applicationVersion.Id,
						Url:         applicationVersion.Url,
						Version:     applicationVersion.Version,
						IsDefault:   applicationVersion.IsDefault,
						Description: applicationVersion.Description,
					})
				case "harmony":
					applicationResponse.Platform.Harmony = append(applicationResponse.Platform.Harmony, &types.ApplicationVersion{
						Id:          applicationVersion.Id,
						Url:         applicationVersion.Url,
						Version:     applicationVersion.Version,
						IsDefault:   applicationVersion.IsDefault,
						Description: applicationVersion.Description,
					})
				}
			}
		}
		resp.Applications = append(resp.Applications, applicationResponse)
	}

	return
}
