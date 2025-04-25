package node

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/app/node"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Get Node list
func GetNodeListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.AppUserSubscbribeNodeRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := node.NewGetNodeListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetNodeList(&req)
		result.HttpResult(c, resp, err)
	}
}
