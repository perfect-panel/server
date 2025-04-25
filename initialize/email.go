package initialize

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perfect-panel/ppanel-server/pkg/logger"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/internal/model/auth"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
)

// Email get email smtp config
func Email(ctx *svc.ServiceContext) {
	logger.Debug("Email config initialization")
	method, err := ctx.AuthModel.FindOneByMethod(context.Background(), "email")
	if err != nil {
		panic(fmt.Sprintf("failed to find email auth method: %v", err.Error()))
	}
	var cfg config.EmailConfig
	var emailConfig = new(auth.EmailAuthConfig)
	if err := emailConfig.Unmarshal(method.Config); err != nil {
		panic(fmt.Sprintf("failed to unmarshal email auth config: %v", err.Error()))
	}
	tool.DeepCopy(&cfg, emailConfig)
	cfg.Enable = *method.Enabled
	value, _ := json.Marshal(emailConfig.PlatformConfig)
	cfg.PlatformConfig = string(value)
	ctx.Config.Email = cfg
}
