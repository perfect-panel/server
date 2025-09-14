package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Query all node tags
func QueryNodeTagHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := server.NewQueryNodeTagLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryNodeTag()
		result.HttpResult(c, resp, err)
	}
}
