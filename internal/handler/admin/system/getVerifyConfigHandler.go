package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get verify config
func GetVerifyConfigHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := system.NewGetVerifyConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetVerifyConfig()
		result.HttpResult(c, resp, err)
	}
}
