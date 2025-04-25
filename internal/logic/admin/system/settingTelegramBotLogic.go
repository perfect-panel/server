package system

import (
	"context"

	"github.com/perfect-panel/ppanel-server/initialize"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
)

type SettingTelegramBotLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSettingTelegramBotLogic setting telegram bot
func NewSettingTelegramBotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SettingTelegramBotLogic {
	return &SettingTelegramBotLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SettingTelegramBotLogic) SettingTelegramBot() error {
	initialize.Telegram(l.svcCtx)
	return nil
}
