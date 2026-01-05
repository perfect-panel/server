package redemption

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/redemption"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Get redemption code list
func GetRedemptionCodeListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetRedemptionCodeListRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := redemption.NewGetRedemptionCodeListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetRedemptionCodeList(&req)
		result.HttpResult(c, resp, err)
	}
}
