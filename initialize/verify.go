package initialize

import (
	"context"

	"github.com/perfect-panel/server/pkg/logger"

	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/tool"
)

type verifyConfig struct {
	TurnstileSiteKey          string
	TurnstileSecret           string
	EnableLoginVerify         bool
	EnableRegisterVerify      bool
	EnableResetPasswordVerify bool
}

func Verify(svc *svc.ServiceContext) {
	logger.Debug("Verify config initialization")
	configs, err := svc.SystemModel.GetVerifyConfig(context.Background())
	if err != nil {
		logger.Error("[Init Verify Config] Get Verify Config Error: ", logger.Field("error", err.Error()))
		return
	}
	var verify verifyConfig
	tool.SystemConfigSliceReflectToStruct(configs, &verify)
	svc.Config.Verify = config.Verify{
		TurnstileSiteKey:    verify.TurnstileSiteKey,
		TurnstileSecret:     verify.TurnstileSecret,
		LoginVerify:         verify.EnableLoginVerify,
		RegisterVerify:      verify.EnableRegisterVerify,
		ResetPasswordVerify: verify.EnableResetPasswordVerify,
	}

	logger.Debug("Verify code config initialization")

	var verifyCodeConfig config.VerifyCode
	cfg, err := svc.SystemModel.GetVerifyCodeConfig(context.Background())
	if err != nil {
		logger.Errorf("[Init Verify Config] Get Verify Code Config Error: %s", err.Error())
		return
	}
	tool.SystemConfigSliceReflectToStruct(cfg, &verifyCodeConfig)
	svc.Config.VerifyCode = verifyCodeConfig
}
