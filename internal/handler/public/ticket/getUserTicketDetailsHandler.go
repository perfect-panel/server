package ticket

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/ticket"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Get ticket detail
func GetUserTicketDetailsHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetUserTicketDetailRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := ticket.NewGetUserTicketDetailsLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetUserTicketDetails(&req)
		result.HttpResult(c, resp, err)
	}
}
