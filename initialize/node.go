package initialize

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/ppanel-server/pkg/logger"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/internal/model/system"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/nodeMultiplier"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
)

func Node(ctx *svc.ServiceContext) {
	logger.Debug("Node config initialization")
	configs, err := ctx.SystemModel.GetNodeConfig(context.Background())
	if err != nil {
		panic(err)
	}
	var nodeConfig config.NodeConfig
	tool.SystemConfigSliceReflectToStruct(configs, &nodeConfig)
	ctx.Config.Node = nodeConfig

	// Manager initialization
	if ctx.DB.Model(&system.System{}).Where("`key` = ?", "NodeMultiplierConfig").Find(&system.System{}).RowsAffected == 0 {
		if err := ctx.DB.Model(&system.System{}).Create(&system.System{
			Key:      "NodeMultiplierConfig",
			Value:    "[]",
			Type:     "string",
			Desc:     "Node Multiplier Config",
			Category: "server",
		}).Error; err != nil {
			logger.Errorf("Create Node Multiplier Config Error: %s", err.Error())
		}
		return
	}

	nodeMultiplierData, err := ctx.SystemModel.FindNodeMultiplierConfig(context.Background())
	if err != nil {

		logger.Error("Get Node Multiplier Config Error: ", logger.Field("error", err.Error()))
		return
	}
	var periods []nodeMultiplier.TimePeriod
	if err := json.Unmarshal([]byte(nodeMultiplierData.Value), &periods); err != nil {
		logger.Error("Unmarshal Node Multiplier Config Error: ", logger.Field("error", err.Error()), logger.Field("value", nodeMultiplierData.Value))
	}
	ctx.NodeMultiplierManager = nodeMultiplier.NewNodeMultiplierManager(periods)
}
