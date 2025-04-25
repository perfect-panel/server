package system

import (
	"context"

	"github.com/perfect-panel/server/internal/model/application"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type CreateApplicationLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateApplicationLogic {
	return &CreateApplicationLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateApplicationLogic) CreateApplication(req *types.CreateApplicationRequest) error {
	var ios []application.ApplicationVersion
	if len(req.Platform.IOS) > 0 {
		for _, ios_ := range req.Platform.IOS {
			ios = append(ios, application.ApplicationVersion{
				Url:         ios_.Url,
				Version:     ios_.Version,
				Platform:    "ios",
				IsDefault:   ios_.IsDefault,
				Description: ios_.Description,
			})
		}
	}

	var mac []application.ApplicationVersion
	if len(req.Platform.MacOS) > 0 {
		for _, mac_ := range req.Platform.MacOS {
			mac = append(mac, application.ApplicationVersion{
				Url:         mac_.Url,
				Version:     mac_.Version,
				Platform:    "macos",
				IsDefault:   mac_.IsDefault,
				Description: mac_.Description,
			})
		}
	}

	var linux []application.ApplicationVersion
	if len(req.Platform.Linux) > 0 {
		for _, linux_ := range req.Platform.Linux {
			linux = append(linux, application.ApplicationVersion{
				Url:         linux_.Url,
				Version:     linux_.Version,
				Platform:    "linux",
				IsDefault:   linux_.IsDefault,
				Description: linux_.Description,
			})
		}
	}

	var android []application.ApplicationVersion
	if len(req.Platform.Android) > 0 {
		for _, android_ := range req.Platform.Android {
			android = append(android, application.ApplicationVersion{
				Url:         android_.Url,
				Version:     android_.Version,
				Platform:    "android",
				IsDefault:   android_.IsDefault,
				Description: android_.Description,
			})
		}
	}

	var windows []application.ApplicationVersion
	if len(req.Platform.Windows) > 0 {
		for _, windows_ := range req.Platform.Windows {
			windows = append(windows, application.ApplicationVersion{
				Url:         windows_.Url,
				Version:     windows_.Version,
				Platform:    "windows",
				IsDefault:   windows_.IsDefault,
				Description: windows_.Description,
			})
		}
	}

	var harmony []application.ApplicationVersion
	if len(req.Platform.Harmony) > 0 {
		for _, harmony_ := range req.Platform.Harmony {
			harmony = append(harmony, application.ApplicationVersion{
				Url:         harmony_.Url,
				Version:     harmony_.Version,
				Platform:    "harmony",
				IsDefault:   harmony_.IsDefault,
				Description: harmony_.Description,
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
	app := application.Application{
		Name:                req.Name,
		Icon:                req.Icon,
		SubscribeType:       req.SubscribeType,
		ApplicationVersions: applicationVersions,
	}
	err := l.svcCtx.ApplicationModel.Insert(l.ctx, &app)
	if err != nil {
		l.Errorw("[CreateApplicationLogic] create application error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create application error: %v", err)
	}
	return nil
}
