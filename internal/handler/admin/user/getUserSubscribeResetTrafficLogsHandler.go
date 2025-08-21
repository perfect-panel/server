package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Get user subcribe reset traffic logs
func GetUserSubscribeResetTrafficLogsHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetUserSubscribeResetTrafficLogsRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := user.NewGetUserSubscribeResetTrafficLogsLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetUserSubscribeResetTrafficLogs(&req)
		result.HttpResult(c, resp, err)
	}
}
