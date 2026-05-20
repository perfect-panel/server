package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get Currency Config
func GetCurrencyConfigHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := system.NewGetCurrencyConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetCurrencyConfig()
		result.HttpResult(c, resp, err)
	}
}
