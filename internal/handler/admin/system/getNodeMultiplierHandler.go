package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get Node Multiplier
func GetNodeMultiplierHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := system.NewGetNodeMultiplierLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetNodeMultiplier()
		result.HttpResult(c, resp, err)
	}
}
