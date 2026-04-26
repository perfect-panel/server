package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// V4.3 我的订阅(含设备槽 + 用量)
func QueryMySubscribesHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		l := subscribe.NewQueryMySubscribesLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryMySubscribes()
		result.HttpResult(c, resp, err)
	}
}
