package console

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/console"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Query revenue statistics
func QueryRevenueStatisticsHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := console.NewQueryRevenueStatisticsLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryRevenueStatistics()
		result.HttpResult(c, resp, err)
	}
}
