package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get Team of Service Config
func GetTosConfigHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := system.NewGetTosConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetTosConfig()
		result.HttpResult(c, resp, err)
	}
}
