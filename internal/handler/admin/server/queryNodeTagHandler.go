package server

import (
	"github.com/perfect-panel/server/internal/logic/admin/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Query all node tags
func QueryNodeTagHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := server.NewQueryNodeTagLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryNodeTag()
		result.HttpResult(c, resp, err)
	}
}
