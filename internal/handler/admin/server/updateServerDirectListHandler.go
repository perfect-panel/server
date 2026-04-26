package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// UpdateServerDirectListHandler Update server direct list.
func UpdateServerDirectListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.UpdateServerDirectListRequest
		_ = c.ShouldBind(&req)
		if id, err := strconv.ParseInt(c.Param("server_id"), 10, 64); err == nil {
			req.ServerId = id
		}
		l := server.NewUpdateServerDirectListLogic(c.Request.Context(), svcCtx)
		resp, err := l.UpdateServerDirectList(&req)
		result.HttpResult(c, resp, err)
	}
}
