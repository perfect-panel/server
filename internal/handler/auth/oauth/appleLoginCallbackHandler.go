package oauth

import (
	"net/http"

	"github.com/perfect-panel/server/internal/logic/auth/oauth"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Apple Login Callback
func AppleLoginCallbackHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.AppleLoginCallbackRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, hertzx.H{"error": "Invalid request data"})
			return
		}
		l := oauth.NewAppleLoginCallbackLogic(c, svcCtx)
		err := l.AppleLoginCallback(&req, c.Request, c.Writer)
		if err != nil {
			result.HttpResult(c, nil, err)
		}
	}
}
