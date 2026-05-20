package user

import (
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get Device List
func GetDeviceListHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := user.NewGetDeviceListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetDeviceList()
		result.HttpResult(c, resp, err)
	}
}
