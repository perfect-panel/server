package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Get alive IP count per uid (aggregated across all nodes)
func GetAliveListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetAliveListRequest
		_ = c.ShouldBindQuery(&req.ServerCommon)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := server.NewGetAliveListLogic(c, svcCtx)
		resp, err := l.GetAliveList(&req)
		if err != nil {
			c.String(500, "Internal Server Error")
			return
		}
		c.JSON(200, resp)
	}
}
