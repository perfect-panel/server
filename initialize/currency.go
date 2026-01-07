package initialize

import (
	"context"
	"fmt"

	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
)

func Currency(ctx *svc.ServiceContext) {
	// Retrieve system currency configuration
	currency, err := ctx.SystemModel.GetCurrencyConfig(context.Background())
	if err != nil {
		logger.Errorf("[INIT] Failed to get currency configuration: %v", err.Error())
		panic(fmt.Sprintf("[INIT] Failed to get currency configuration: %v", err.Error()))
	}
	// Parse currency configuration
	configs := struct {
		CurrencyUnit   string
		CurrencySymbol string
		AccessKey      string
	}{}
	tool.SystemConfigSliceReflectToStruct(currency, &configs)
	ctx.ExchangeRate = 0 // Default exchange rate to 0
	ctx.Config.Currency = config.Currency{
		Unit:      configs.CurrencyUnit,
		Symbol:    configs.CurrencySymbol,
		AccessKey: configs.AccessKey,
	}
	logger.Infof("[INIT] Currency configuration: %v", ctx.Config.Currency)
}
