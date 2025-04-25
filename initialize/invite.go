package initialize

import (
	"context"

	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
)

func Invite(ctx *svc.ServiceContext) {
	// Initialize the system configuration
	logger.Debug("Register config initialization")
	configs, err := ctx.SystemModel.GetInviteConfig(context.Background())
	if err != nil {
		logger.Error("[Init Invite Config] Get Invite Config Error: ", logger.Field("error", err.Error()))
		return
	}
	var inviteConfig config.InviteConfig
	tool.SystemConfigSliceReflectToStruct(configs, &inviteConfig)
	ctx.Config.Invite = inviteConfig
}
