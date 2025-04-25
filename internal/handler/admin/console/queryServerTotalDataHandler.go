package console

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/console"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Query server total data
func QueryServerTotalDataHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := console.NewQueryServerTotalDataLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryServerTotalData()
		result.HttpResult(c, resp, err)
	}
}
