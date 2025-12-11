package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Reset all subscribe tokens
func ResetAllSubscribeTokenHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := subscribe.NewResetAllSubscribeTokenLogic(c.Request.Context(), svcCtx)
		resp, err := l.ResetAllSubscribeToken()
		result.HttpResult(c, resp, err)
	}
}
