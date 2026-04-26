package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Get limiter online status for a single user (online IPs + 24h reject count)
func GetUserOnlineStatusHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetUserOnlineStatusRequest
		_ = c.ShouldBindQuery(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := user.NewGetUserOnlineStatusLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetUserOnlineStatus(&req)
		result.HttpResult(c, resp, err)
	}
}
