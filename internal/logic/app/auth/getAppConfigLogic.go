package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/perfect-panel/server/internal/model/application"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type GetAppConfigLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// GetAppConfig
func NewGetAppConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAppConfigLogic {
	return &GetAppConfigLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAppConfigLogic) GetAppConfig(req *types.AppConfigRequest) (resp *types.AppConfigResponse, err error) {
	resp = &types.AppConfigResponse{}
	systems, err := l.svcCtx.SystemModel.GetSiteConfig(l.ctx)
	if err != nil {
		l.Errorw("[QueryApplicationConfig] GetSiteConfig error: ", logger.Field("error", err.Error()))
	}
	for _, sysVal := range systems {
		if sysVal.Key == "CustomData" {
			jsonStr := strings.ReplaceAll(sysVal.Value, "\\", "")
			customData := make(map[string]interface{})
			if err = json.Unmarshal([]byte(jsonStr), &customData); err != nil {
				break
			}

			website := customData["website"]
			if website != nil {
				resp.OfficialWebsite = fmt.Sprintf("%v", website)
			}

			contacts := customData["contacts"]
			if contacts != nil {
				contactsJson, err := json.Marshal(contacts)
				if err == nil {
					contactsMap := make(map[string]string)
					err = json.Unmarshal(contactsJson, &contactsMap)
					if err == nil {
						resp.OfficialEmail = fmt.Sprintf("%v", contactsMap["email"])
						resp.OfficialTelegram = fmt.Sprintf("%v", contactsMap["telegram"])
						resp.OfficialTelephone = fmt.Sprintf("%v", contactsMap["telephone"])
					}
				}
			}
			break
		}
	}

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

	isOk := false
	for _, app := range applications {
		if isOk {
			break
		}
		resp.Application.Name = app.Name
		resp.Application.Description = app.Description
		applicationVersions := app.ApplicationVersions
		if len(applicationVersions) != 0 {
			for _, applicationVersion := range applicationVersions {
				if applicationVersion.Platform == req.UserAgent {
					resp.Application.Id = applicationVersion.ApplicationId
					resp.Application.Url = applicationVersion.Url
					resp.Application.Version = applicationVersion.Version
					resp.Application.VersionDescription = applicationVersion.Description
					resp.Application.IsDefault = applicationVersion.IsDefault
					isOk = true
					break
				}
			}
		}
	}

	configs, err := l.svcCtx.ApplicationModel.FindOneConfig(l.ctx, 1)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Logger.Error("[GetAppInfo] FindOneAppConfig error: ", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "GetAppInfo FindOneAppConfig error: %v", err.Error())
	}
	resp.EncryptionKey = configs.EncryptionKey
	resp.EncryptionMethod = configs.EncryptionMethod
	resp.Domains = strings.Split(configs.Domains, ";")
	resp.StartupPicture = configs.StartupPicture
	resp.StartupPictureSkipTime = configs.StartupPictureSkipTime
	resp.InvitationLink = configs.InvitationLink
	resp.KrWebsiteId = configs.KrWebsiteId
	return
}
