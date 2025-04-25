package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get subscribe group list
func QuerySubscribeGroupListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := subscribe.NewQuerySubscribeGroupListLogic(c.Request.Context(), svcCtx)
		resp, err := l.QuerySubscribeGroupList()
		result.HttpResult(c, resp, err)
	}
}
