package server_bak

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/server_bak"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get node tag list
func GetNodeTagListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := server_bak.NewGetNodeTagListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetNodeTagList()
		result.HttpResult(c, resp, err)
	}
}
