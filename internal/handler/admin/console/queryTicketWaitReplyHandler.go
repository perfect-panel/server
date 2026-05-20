package console

import (
	"github.com/perfect-panel/server/internal/logic/admin/console"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Query ticket wait reply
func QueryTicketWaitReplyHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := console.NewQueryTicketWaitReplyLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryTicketWaitReply()
		result.HttpResult(c, resp, err)
	}
}
