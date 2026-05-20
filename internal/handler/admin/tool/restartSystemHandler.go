package tool

import (
	"github.com/perfect-panel/server/internal/logic/admin/tool"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Restart System
func RestartSystemHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := tool.NewRestartSystemLogic(c.Request.Context(), svcCtx)
		err := l.RestartSystem()
		result.HttpResult(c, nil, err)
	}
}
