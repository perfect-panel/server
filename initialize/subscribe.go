package initialize

import (
	"context"

	"github.com/perfect-panel/ppanel-server/pkg/logger"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
)

func Subscribe(svc *svc.ServiceContext) {
	logger.Debug("Subscribe config initialization")
	configs, err := svc.SystemModel.GetSubscribeConfig(context.Background())
	if err != nil {
		logger.Error("[Init Subscribe Config] Get Subscribe Config Error: ", logger.Field("error", err.Error()))
		return
	}

	var subscribeConfig config.SubscribeConfig
	tool.SystemConfigSliceReflectToStruct(configs, &subscribeConfig)
	svc.Config.Subscribe = subscribeConfig
}
