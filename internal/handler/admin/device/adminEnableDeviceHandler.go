package device

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/device"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// AdminEnableDeviceHandler Force-enable a device slot.
// V4.4 #44: admin acts on user devices (audit actor=admin).
func AdminEnableDeviceHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.AdminDeviceIdRequest
		if id, err := strconv.ParseInt(c.Param("device_id"), 10, 64); err == nil {
			req.DeviceId = id
		}
		l := device.NewAdminEnableDeviceLogic(c, svcCtx)
		resp, err := l.AdminEnableDevice(&req)
		result.HttpResult(c, resp, err)
	}
}
