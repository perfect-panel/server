package node

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/app/node"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get rule group list
func GetRuleGroupListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := node.NewGetRuleGroupListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetRuleGroupList()
		result.HttpResult(c, resp, err)
	}
}
