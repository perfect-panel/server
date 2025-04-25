package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/app/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Get user online time total
func GetUserOnlineTimeStatisticsHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := user.NewGetUserOnlineTimeStatisticsLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetUserOnlineTimeStatistics()
		result.HttpResult(c, resp, err)
	}
}
