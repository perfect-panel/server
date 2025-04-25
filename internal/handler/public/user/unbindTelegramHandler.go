package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Unbind Telegram
func UnbindTelegramHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := user.NewUnbindTelegramLogic(c.Request.Context(), svcCtx)
		err := l.UnbindTelegram()
		result.HttpResult(c, nil, err)
	}
}
