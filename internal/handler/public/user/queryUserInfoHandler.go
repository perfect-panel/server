package user

import (
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Query User Info
func QueryUserInfoHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := user.NewQueryUserInfoLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryUserInfo()
		result.HttpResult(c, resp, err)
	}
}
