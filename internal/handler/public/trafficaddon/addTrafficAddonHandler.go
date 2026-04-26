package trafficaddon

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/trafficaddon"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// V4.3 流量加购
func AddTrafficAddonHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.AddTrafficAddonRequest
		_ = c.ShouldBind(&req)
		if id, err := strconv.ParseInt(c.Param("id"), 10, 64); err == nil {
			req.UserSubscribeId = id
		}
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}
		l := trafficaddon.NewAddTrafficAddonLogic(c.Request.Context(), svcCtx)
		resp, err := l.AddTrafficAddon(&req)
		result.HttpResult(c, resp, err)
	}
}
