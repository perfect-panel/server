package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/public/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Query User Balance Log
func QueryUserBalanceLogHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := user.NewQueryUserBalanceLogLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryUserBalanceLog()
		result.HttpResult(c, resp, err)
	}
}
