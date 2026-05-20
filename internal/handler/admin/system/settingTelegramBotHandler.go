package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// setting telegram bot
func SettingTelegramBotHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := system.NewSettingTelegramBotLogic(c.Request.Context(), svcCtx)
		err := l.SettingTelegramBot()
		result.HttpResult(c, nil, err)
	}
}
