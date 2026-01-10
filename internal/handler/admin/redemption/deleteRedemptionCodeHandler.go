package redemption

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/redemption"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Delete redemption code
func DeleteRedemptionCodeHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.DeleteRedemptionCodeRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := redemption.NewDeleteRedemptionCodeLogic(c.Request.Context(), svcCtx)
		err := l.DeleteRedemptionCode(&req)
		result.HttpResult(c, nil, err)
	}
}
