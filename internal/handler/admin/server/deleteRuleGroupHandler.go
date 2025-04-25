package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/admin/server"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Delete rule group
func DeleteRuleGroupHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.DeleteRuleGroupRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := server.NewDeleteRuleGroupLogic(c.Request.Context(), svcCtx)
		err := l.DeleteRuleGroup(&req)
		result.HttpResult(c, nil, err)
	}
}
