package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/app/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get application config
func QueryApplicationConfigHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := subscribe.NewQueryApplicationConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryApplicationConfig()
		result.HttpResult(c, resp, err)
	}
}
