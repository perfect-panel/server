package device

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/device"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// V4.3 加购单设备 / 多设备(quantity 在 body 里)
func AddSubscribeDeviceHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.AddSubscribeDeviceRequest
		// 先读 JSON body(quantity 等可选字段),再用 path 覆盖 id —
		// 不能先 path 再 body,否则 ShouldBindJSON 会把 id 清零。
		_ = c.ShouldBindJSON(&req)
		if id, err := strconv.ParseInt(c.Param("id"), 10, 64); err == nil {
			req.UserSubscribeId = id
		}
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}
		l := device.NewAddSubscribeDeviceLogic(c.Request.Context(), svcCtx)
		resp, err := l.AddSubscribeDevice(&req)
		result.HttpResult(c, resp, err)
	}
}
