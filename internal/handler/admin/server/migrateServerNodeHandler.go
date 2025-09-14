package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Migrate server and node data to new database
func MigrateServerNodeHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := server.NewMigrateServerNodeLogic(c.Request.Context(), svcCtx)
		resp, err := l.MigrateServerNode()
		result.HttpResult(c, resp, err)
	}
}
