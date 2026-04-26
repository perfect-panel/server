package device

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/device"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// AdminRenameDeviceHandler Rename a device.
func AdminRenameDeviceHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.AdminDeviceRenameRequest
		_ = c.ShouldBind(&req)
		if id, err := strconv.ParseInt(c.Param("device_id"), 10, 64); err == nil {
			req.DeviceId = id
		}
		l := device.NewAdminRenameDeviceLogic(c, svcCtx)
		resp, err := l.AdminRenameDevice(&req)
		result.HttpResult(c, resp, err)
	}
}
