package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get subscribe group list
func GetSubscribeGroupListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := subscribe.NewGetSubscribeGroupListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetSubscribeGroupList()
		result.HttpResult(c, resp, err)
	}
}
