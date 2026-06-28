package authMethod

import (
	"github.com/perfect-panel/server/internal/logic/admin/authMethod"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get auth method list
func GetAuthMethodListHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := authMethod.NewGetAuthMethodListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetAuthMethodList()
		result.HttpResult(c, resp, err)
	}
}
