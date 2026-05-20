package tool

import (
	"github.com/perfect-panel/server/internal/logic/admin/tool"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// GetVersionHandler Get Version
func GetVersionHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := tool.NewGetVersionLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetVersion()
		result.HttpResult(c, resp, err)
	}
}
