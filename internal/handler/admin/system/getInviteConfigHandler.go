package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get invite config
func GetInviteConfigHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := system.NewGetInviteConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetInviteConfig()
		result.HttpResult(c, resp, err)
	}
}
