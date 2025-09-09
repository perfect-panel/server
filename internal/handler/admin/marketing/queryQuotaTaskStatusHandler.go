package marketing

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/marketing"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Query quota task status
func QueryQuotaTaskStatusHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.QueryQuotaTaskStatusRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := marketing.NewQueryQuotaTaskStatusLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryQuotaTaskStatus(&req)
		result.HttpResult(c, resp, err)
	}
}
