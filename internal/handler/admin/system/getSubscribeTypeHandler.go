package system

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/admin/system"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Get subscribe type
func GetSubscribeTypeHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := system.NewGetSubscribeTypeLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetSubscribeType()
		result.HttpResult(c, resp, err)
	}
}
