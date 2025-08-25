package server_bak

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/server_bak"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get node group list
func GetNodeGroupListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := server_bak.NewGetNodeGroupListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetNodeGroupList()
		result.HttpResult(c, resp, err)
	}
}
