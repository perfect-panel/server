package system

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get subscribe config
func GetSubscribeConfigHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := system.NewGetSubscribeConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetSubscribeConfig()
		result.HttpResult(c, resp, err)
	}
}
