package user

import (
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get OAuth Methods
func GetOAuthMethodsHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := user.NewGetOAuthMethodsLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetOAuthMethods()
		result.HttpResult(c, resp, err)
	}
}
