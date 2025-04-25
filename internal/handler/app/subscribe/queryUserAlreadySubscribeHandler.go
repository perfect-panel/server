package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/app/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get  Already subscribed to package
func QueryUserAlreadySubscribeHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := subscribe.NewQueryUserAlreadySubscribeLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryUserAlreadySubscribe()
		result.HttpResult(c, resp, err)
	}
}
