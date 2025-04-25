package system

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get Node Multiplier
func GetNodeMultiplierHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := system.NewGetNodeMultiplierLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetNodeMultiplier()
		result.HttpResult(c, resp, err)
	}
}
