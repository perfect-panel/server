package system

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// GetModuleConfigHandler Get Module Config
func GetModuleConfigHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := system.NewGetModuleConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetModuleConfig()
		result.HttpResult(c, resp, err)
	}
}
