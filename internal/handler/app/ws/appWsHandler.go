package ws

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/app/ws"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// App heartbeat
func AppWsHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Logic: App heartbeat
		l := ws.NewAppWsLogic(ctx, svcCtx)
		err := l.AppWs(c.Writer, c.Request, c.Param("userid"), c.Param("identifier"))
		result.HttpResult(c, nil, err)
	}
}
