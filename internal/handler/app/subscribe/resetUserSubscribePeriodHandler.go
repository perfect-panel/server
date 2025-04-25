package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/app/subscribe"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Reset user subscription period
func ResetUserSubscribePeriodHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.UserSubscribeResetPeriodRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := subscribe.NewResetUserSubscribePeriodLogic(c.Request.Context(), svcCtx)
		resp, err := l.ResetUserSubscribePeriod(&req)
		result.HttpResult(c, resp, err)
	}
}
