package initialize

import (
	"context"

	"github.com/perfect-panel/ppanel-server/pkg/logger"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
)

func Site(ctx *svc.ServiceContext) {
	logger.Debug("initialize site config")
	configs, err := ctx.SystemModel.GetSiteConfig(context.Background())
	if err != nil {
		panic(err)
	}
	var siteConfig config.SiteConfig
	tool.SystemConfigSliceReflectToStruct(configs, &siteConfig)
	ctx.Config.Site = siteConfig
}
