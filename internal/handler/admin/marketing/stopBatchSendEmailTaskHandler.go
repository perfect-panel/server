package marketing

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/marketing"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// StopBatchSendEmailTaskHandler Stop a batch send email task
func StopBatchSendEmailTaskHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.StopBatchSendEmailTaskRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := marketing.NewStopBatchSendEmailTaskLogic(c.Request.Context(), svcCtx)
		err := l.StopBatchSendEmailTask(&req)
		result.HttpResult(c, nil, err)
	}
}
