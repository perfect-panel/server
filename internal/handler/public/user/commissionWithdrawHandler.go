package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Commission Withdraw
func CommissionWithdrawHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.CommissionWithdrawRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := user.NewCommissionWithdrawLogic(c.Request.Context(), svcCtx)
		resp, err := l.CommissionWithdraw(&req)
		result.HttpResult(c, resp, err)
	}
}
