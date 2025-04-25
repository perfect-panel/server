package system

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/admin/system"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// get application config
func GetApplicationConfigHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := system.NewGetApplicationConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetApplicationConfig()
		result.HttpResult(c, resp, err)
	}
}
