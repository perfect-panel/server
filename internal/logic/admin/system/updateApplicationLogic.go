package system

import (
	"context"

	"github.com/perfect-panel/server/internal/model/application"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
)

type UpdateApplicationLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateApplicationLogic {
	return &UpdateApplicationLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateApplicationLogic) UpdateApplication(req *types.UpdateApplicationRequest) error {

	// find application
	app, err := l.svcCtx.ApplicationModel.FindOne(l.ctx, req.Id)
	if err != nil {
		l.Errorw("[UpdateApplication] find application error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find application error: %v", err.Error())
	}
	app.Name = req.Name
	app.Icon = req.Icon
	app.SubscribeType = req.SubscribeType
	app.Description = req.Description

	var ios []application.ApplicationVersion
	if len(req.Platform.IOS) > 0 {
		for _, ios_ := range req.Platform.IOS {
			ios = append(ios, application.ApplicationVersion{
				Url:           ios_.Url,
				Version:       ios_.Version,
				Platform:      "ios",
				IsDefault:     ios_.IsDefault,
				Description:   ios_.Description,
				ApplicationId: app.Id,
			})
		}
	}

	var mac []application.ApplicationVersion
	if len(req.Platform.MacOS) > 0 {
		for _, mac_ := range req.Platform.MacOS {
			mac = append(mac, application.ApplicationVersion{
				Url:           mac_.Url,
				Version:       mac_.Version,
				Platform:      "macos",
				IsDefault:     mac_.IsDefault,
				Description:   mac_.Description,
				ApplicationId: app.Id,
			})
		}
	}

	var linux []application.ApplicationVersion
	if len(req.Platform.Linux) > 0 {
		for _, linux_ := range req.Platform.Linux {
			linux = append(linux, application.ApplicationVersion{
				Url:           linux_.Url,
				Version:       linux_.Version,
				Platform:      "linux",
				IsDefault:     linux_.IsDefault,
				Description:   linux_.Description,
				ApplicationId: app.Id,
			})
		}
	}

	var android []application.ApplicationVersion
	if len(req.Platform.Android) > 0 {
		for _, android_ := range req.Platform.Android {
			android = append(android, application.ApplicationVersion{
				Url:           android_.Url,
				Version:       android_.Version,
				Platform:      "android",
				IsDefault:     android_.IsDefault,
				Description:   android_.Description,
				ApplicationId: app.Id,
			})
		}
	}

	var windows []application.ApplicationVersion
	if len(req.Platform.Windows) > 0 {
		for _, windows_ := range req.Platform.Windows {
			windows = append(windows, application.ApplicationVersion{
				Url:           windows_.Url,
				Version:       windows_.Version,
				Platform:      "windows",
				IsDefault:     windows_.IsDefault,
				Description:   windows_.Description,
				ApplicationId: app.Id,
			})
		}
	}

	var harmony []application.ApplicationVersion
	if len(req.Platform.Harmony) > 0 {
		for _, harmony_ := range req.Platform.Harmony {
			harmony = append(harmony, application.ApplicationVersion{
				Url:           harmony_.Url,
				Version:       harmony_.Version,
				Platform:      "harmony",
				IsDefault:     harmony_.IsDefault,
				Description:   harmony_.Description,
				ApplicationId: app.Id,
			})
		}
	}
	var applicationVersions []application.ApplicationVersion
	applicationVersions = append(applicationVersions, ios...)
	applicationVersions = append(applicationVersions, mac...)
	applicationVersions = append(applicationVersions, linux...)
	applicationVersions = append(applicationVersions, android...)
	applicationVersions = append(applicationVersions, windows...)
	applicationVersions = append(applicationVersions, harmony...)
	app.ApplicationVersions = applicationVersions
	err = l.svcCtx.ApplicationModel.Transaction(l.ctx, func(db *gorm.DB) error {

		if err = db.Where("application_id = ?", app.Id).Delete(&application.ApplicationVersion{}).Error; err != nil {
			return err
		}
		if err = db.Create(&applicationVersions).Error; err != nil {
			return err
		}
		return db.Save(app).Error
	})
	if err != nil {
		l.Errorw("[UpdateApplication] update application error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update application error: %v", err.Error())
	}
	return nil
}
