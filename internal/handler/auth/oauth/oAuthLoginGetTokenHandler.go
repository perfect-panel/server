package oauth

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/auth/oauth"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// OAuth login get token
func OAuthLoginGetTokenHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.OAuthLoginGetTokenRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := oauth.NewOAuthLoginGetTokenLogic(c.Request.Context(), svcCtx)
		resp, err := l.OAuthLoginGetToken(&req, c.ClientIP(), c.Request.UserAgent())
		result.HttpResult(c, resp, err)
	}
}
