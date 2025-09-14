package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Check if there is any server or node to migrate
func HasMigrateSeverNodeHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := server.NewHasMigrateSeverNodeLogic(c.Request.Context(), svcCtx)
		resp, err := l.HasMigrateSeverNode()
		result.HttpResult(c, resp, err)
	}
}
