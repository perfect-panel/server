package user

import (
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Unbind Telegram
func UnbindTelegramHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := user.NewUnbindTelegramLogic(c.Request.Context(), svcCtx)
		err := l.UnbindTelegram()
		result.HttpResult(c, nil, err)
	}
}
