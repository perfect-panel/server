package redemption

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/redemption"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Toggle redemption code status
func ToggleRedemptionCodeStatusHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.ToggleRedemptionCodeStatusRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := redemption.NewToggleRedemptionCodeStatusLogic(c.Request.Context(), svcCtx)
		err := l.ToggleRedemptionCodeStatus(&req)
		result.HttpResult(c, nil, err)
	}
}
