package ticket

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/ticket"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Get ticket list
func GetUserTicketListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetUserTicketListRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := ticket.NewGetUserTicketListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetUserTicketList(&req)
		result.HttpResult(c, resp, err)
	}
}
