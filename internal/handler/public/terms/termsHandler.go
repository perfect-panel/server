package terms

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/terms"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

func StatusHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		l := terms.NewStatusLogic(c, svcCtx)
		resp, err := l.Status()
		result.HttpResult(c, resp, err)
	}
}

func AcceptHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		l := terms.NewAcceptLogic(c, svcCtx)
		resp, err := l.Accept()
		result.HttpResult(c, resp, err)
	}
}
