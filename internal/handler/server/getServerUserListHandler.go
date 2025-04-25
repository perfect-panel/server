package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

// Get user list
func GetServerUserListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetServerUserListRequest
		_ = c.ShouldBind(&req)
		_ = c.ShouldBindQuery(&req.ServerCommon)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := server.NewGetServerUserListLogic(c, svcCtx)
		resp, err := l.GetServerUserList(&req)
		if err != nil {
			if errors.Is(err, xerr.StatusNotModified) {
				c.String(304, "Not Modified")
				return
			}
			c.String(404, "Not Found")
			return
		}
		c.JSON(200, resp)
	}
}
