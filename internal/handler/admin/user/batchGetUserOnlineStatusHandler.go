package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Batch get limiter online status (list-page augmentation)
func BatchGetUserOnlineStatusHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.BatchGetUserOnlineStatusRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := user.NewBatchGetUserOnlineStatusLogic(c.Request.Context(), svcCtx)
		resp, err := l.BatchGetUserOnlineStatus(&req)
		result.HttpResult(c, resp, err)
	}
}
