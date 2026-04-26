package device

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/device"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// AdminResetDeviceHandler Force-reset a device token.
func AdminResetDeviceHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.AdminDeviceIdRequest
		if id, err := strconv.ParseInt(c.Param("device_id"), 10, 64); err == nil {
			req.DeviceId = id
		}
		l := device.NewAdminResetDeviceLogic(c, svcCtx)
		resp, err := l.AdminResetDevice(&req)
		result.HttpResult(c, resp, err)
	}
}
