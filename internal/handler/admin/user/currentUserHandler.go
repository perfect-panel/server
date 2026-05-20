package user

import (
	"github.com/perfect-panel/server/internal/logic/admin/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Current user
func CurrentUserHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		l := user.NewCurrentUserLogic(c.Request.Context(), svcCtx)
		resp, err := l.CurrentUser()
		result.HttpResult(c, resp, err)
	}
}
