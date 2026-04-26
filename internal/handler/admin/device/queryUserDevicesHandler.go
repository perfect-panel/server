package device

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/device"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// QueryUserDevicesHandler Query a user's device slots.
func QueryUserDevicesHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.QueryUserDevicesRequest
		_ = c.ShouldBindQuery(&req)
		l := device.NewQueryUserDevicesLogic(c, svcCtx)
		resp, err := l.QueryUserDevices(&req)
		result.HttpResult(c, resp, err)
	}
}
