package order

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/order"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Get order list
func QueryOrderListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.QueryOrderListRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := order.NewQueryOrderListLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryOrderList(&req)
		result.HttpResult(c, resp, err)
	}
}
