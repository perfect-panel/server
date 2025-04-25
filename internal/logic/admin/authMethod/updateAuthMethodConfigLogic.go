package authMethod

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/server/initialize"
	"github.com/perfect-panel/server/internal/model/auth"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/email"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/sms"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdateAuthMethodConfigLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update auth method config
func NewUpdateAuthMethodConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAuthMethodConfigLogic {
	return &UpdateAuthMethodConfigLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAuthMethodConfigLogic) UpdateAuthMethodConfig(req *types.UpdateAuthMethodConfigRequest) (resp *types.AuthMethodConfig, err error) {
	method, err := l.svcCtx.AuthModel.FindOneByMethod(l.ctx, req.Method)
	if err != nil {
		l.Errorw("find one by method failed", logger.Field("method", req.Method), logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find one by method failed: %v", err.Error())
	}

	tool.DeepCopy(method, req)
	if req.Config != nil {
		if value, ok := req.Config.(map[string]interface{}); ok {
			if req.Method == "email" && value["verify_email_template"] == "" {
				value["verify_email_template"] = email.DefaultEmailVerifyTemplate
			}
			if req.Method == "email" && value["expiration_email_template"] == "" {
				value["expiration_email_template"] = email.DefaultExpirationEmailTemplate
			}
			if req.Method == "email" && value["maintenance_email_template"] == "" {
				value["maintenance_email_template"] = email.DefaultMaintenanceEmailTemplate
			}
			if req.Method == "email" && value["traffic_exceed_email_template"] == "" {
				value["traffic_exceed_email_template"] = email.DefaultTrafficExceedEmailTemplate
			}

			if value["platform_config"] != nil {
				platformConfig, err := validatePlatformConfig(value["platform"].(string), value["platform_config"].(map[string]interface{}))
				if err != nil {
					l.Errorw("validate platform config failed", logger.Field("config", req.Config), logger.Field("error", err.Error()))
					return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "validate platform config failed: %v", err.Error())
				}
				req.Config.(map[string]interface{})["platform_config"] = platformConfig
			}
		}
		bytes, err := json.Marshal(req.Config)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "marshal config failed: %v", err.Error())
		}
		method.Config = string(bytes)
	}
	err = l.svcCtx.AuthModel.Update(l.ctx, method)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "update auth method failed: %v", err.Error())
	}

	resp = new(types.AuthMethodConfig)
	tool.DeepCopy(resp, method)
	if method.Config != "" {
		if err := json.Unmarshal([]byte(method.Config), &resp.Config); err != nil {
			l.Errorw("unmarshal apple config failed", logger.Field("config", method.Config), logger.Field("error", err.Error()))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "unmarshal apple config failed: %v", err.Error())
		}
	}
	// update global config
	defer l.UpdateGlobal(method.Method)
	return
}

func (l *UpdateAuthMethodConfigLogic) UpdateGlobal(method string) {
	if method == "email" {
		initialize.Email(l.svcCtx)
	}
	if method == "mobile" {
		initialize.Mobile(l.svcCtx)
	}
}

func validatePlatformConfig(platform string, cfg map[string]interface{}) (interface{}, error) {
	var bytes []byte
	var err error
	var config interface{}
	bytes, err = json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	switch platform {
	case email.SMTP.String():
		config = new(auth.SMTPConfig)
	case sms.Abosend.String():
		config = new(auth.AbosendConfig)
	case sms.AlibabaCloud.String():
		config = new(auth.AlibabaCloudConfig)
	case sms.Smsbao.String():
		config = new(auth.SmsbaoConfig)
	case sms.Twilio.String():
		config = new(auth.TwilioConfig)
	default:
		return nil, errors.New("invalid platform")
	}
	err = json.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
