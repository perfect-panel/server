package device

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/device"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

func parseDeviceId(c *gin.Context, req *types.DeviceIdRequest) {
	if id, err := strconv.ParseInt(c.Param("device_id"), 10, 64); err == nil {
		req.DeviceId = id
	}
}

// V4.3 重置设备
func ResetDeviceHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.DeviceIdRequest
		parseDeviceId(c, &req)
		l := device.NewResetDeviceLogic(c.Request.Context(), svcCtx)
		resp, err := l.ResetDevice(&req)
		result.HttpResult(c, resp, err)
	}
}

// V4.3 停用设备
func DisableDeviceHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.DeviceIdRequest
		parseDeviceId(c, &req)
		l := device.NewDisableDeviceLogic(c.Request.Context(), svcCtx)
		resp, err := l.DisableDevice(&req)
		result.HttpResult(c, resp, err)
	}
}

// V4.3 启用设备(并换发新 token+uuid)
func EnableDeviceHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.DeviceIdRequest
		parseDeviceId(c, &req)
		l := device.NewEnableDeviceLogic(c.Request.Context(), svcCtx)
		resp, err := l.EnableDevice(&req)
		result.HttpResult(c, resp, err)
	}
}

// V4.3 改设备名
func RenameDeviceHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.DeviceRenameRequest
		_ = c.ShouldBind(&req)
		if id, err := strconv.ParseInt(c.Param("device_id"), 10, 64); err == nil {
			req.DeviceId = id
		}
		l := device.NewRenameDeviceLogic(c.Request.Context(), svcCtx)
		resp, err := l.RenameDevice(&req)
		result.HttpResult(c, resp, err)
	}
}

// V4.3 一键重置全部
func ResetAllDevicesHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.ResetAllDevicesRequest
		if id, err := strconv.ParseInt(c.Param("id"), 10, 64); err == nil {
			req.UserSubscribeId = id
		}
		l := device.NewResetAllDevicesLogic(c.Request.Context(), svcCtx)
		resp, err := l.ResetAllDevices(&req)
		result.HttpResult(c, resp, err)
	}
}

// V4.3 删除加购设备(仅 is_addon=true 可删,无退款)
func DeleteAddonDeviceHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.DeviceIdRequest
		parseDeviceId(c, &req)
		l := device.NewDeleteAddonDeviceLogic(c.Request.Context(), svcCtx)
		resp, err := l.DeleteAddonDevice(&req)
		result.HttpResult(c, resp, err)
	}
}
