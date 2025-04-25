package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/app/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Get user subcribe traffic logs
func GetUserSubscribeTrafficLogsHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetUserSubscribeTrafficLogsRequest
		_ = c.BindQuery(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := user.NewGetUserSubscribeTrafficLogsLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetUserSubscribeTrafficLogs(&req)
		result.HttpResult(c, resp, err)
	}
}
