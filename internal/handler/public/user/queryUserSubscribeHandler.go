package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Query User Subscribe
func QueryUserSubscribeHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := user.NewQueryUserSubscribeLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryUserSubscribe()
		result.HttpResult(c, resp, err)
	}
}
