package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/admin/server"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Get node group list
func GetNodeGroupListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := server.NewGetNodeGroupListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetNodeGroupList()
		result.HttpResult(c, resp, err)
	}
}
