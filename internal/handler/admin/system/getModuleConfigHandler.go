package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// GetModuleConfigHandler Get Module Config
func GetModuleConfigHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := system.NewGetModuleConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetModuleConfig()
		result.HttpResult(c, resp, err)
	}
}
