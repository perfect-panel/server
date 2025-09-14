package log

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/log"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Filter user subscribe traffic log
func FilterUserSubscribeTrafficLogHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.FilterSubscribeTrafficRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := log.NewFilterUserSubscribeTrafficLogLogic(c.Request.Context(), svcCtx)
		resp, err := l.FilterUserSubscribeTrafficLog(&req)
		result.HttpResult(c, resp, err)
	}
}
