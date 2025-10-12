package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get Device List
func GetDeviceListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := user.NewGetDeviceListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetDeviceList()
		result.HttpResult(c, resp, err)
	}
}
