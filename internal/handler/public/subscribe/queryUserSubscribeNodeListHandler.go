package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get user subscribe node info
func QueryUserSubscribeNodeListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := subscribe.NewQueryUserSubscribeNodeListLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryUserSubscribeNodeList()
		result.HttpResult(c, resp, err)
	}
}
