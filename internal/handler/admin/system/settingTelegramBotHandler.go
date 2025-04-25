package system

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/admin/system"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// setting telegram bot
func SettingTelegramBotHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := system.NewSettingTelegramBotLogic(c.Request.Context(), svcCtx)
		err := l.SettingTelegramBot()
		result.HttpResult(c, nil, err)
	}
}
