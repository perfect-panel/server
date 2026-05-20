package ticket

import (
	"github.com/perfect-panel/server/internal/logic/public/ticket"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Create ticket follow
func CreateUserTicketFollowHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.CreateUserTicketFollowRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := ticket.NewCreateUserTicketFollowLogic(c.Request.Context(), svcCtx)
		err := l.CreateUserTicketFollow(&req)
		result.HttpResult(c, nil, err)
	}
}
