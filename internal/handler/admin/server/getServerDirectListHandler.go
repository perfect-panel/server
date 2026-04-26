package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// GetServerDirectListHandler Get server direct list.
func GetServerDirectListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetServerDirectListRequest
		if id, err := strconv.ParseInt(c.Param("server_id"), 10, 64); err == nil {
			req.ServerId = id
		}
		l := server.NewGetServerDirectListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetServerDirectList(&req)
		result.HttpResult(c, resp, err)
	}
}
