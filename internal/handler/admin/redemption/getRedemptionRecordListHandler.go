package redemption

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/redemption"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Get redemption record list
func GetRedemptionRecordListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetRedemptionRecordListRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := redemption.NewGetRedemptionRecordListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetRedemptionRecordList(&req)
		result.HttpResult(c, resp, err)
	}
}
