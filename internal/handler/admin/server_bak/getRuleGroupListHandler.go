package server_bak

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/server_bak"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get rule group list
func GetRuleGroupListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := server_bak.NewGetRuleGroupListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetRuleGroupList()
		result.HttpResult(c, resp, err)
	}
}
