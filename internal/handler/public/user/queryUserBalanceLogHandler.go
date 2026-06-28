package user

import (
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Query User Balance Log
func QueryUserBalanceLogHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := user.NewQueryUserBalanceLogLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryUserBalanceLog()
		result.HttpResult(c, resp, err)
	}
}
