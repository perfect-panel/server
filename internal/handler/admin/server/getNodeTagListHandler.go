package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/admin/server"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Get node tag list
func GetNodeTagListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := server.NewGetNodeTagListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetNodeTagList()
		result.HttpResult(c, resp, err)
	}
}
