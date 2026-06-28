package tool

import (
	"github.com/perfect-panel/server/internal/logic/admin/tool"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get System Log
func GetSystemLogHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := tool.NewGetSystemLogLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetSystemLog()
		result.HttpResult(c, resp, err)
	}
}
