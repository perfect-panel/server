package user

import (
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Bind Telegram
func BindTelegramHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := user.NewBindTelegramLogic(c.Request.Context(), svcCtx)
		resp, err := l.BindTelegram()
		result.HttpResult(c, resp, err)
	}
}
