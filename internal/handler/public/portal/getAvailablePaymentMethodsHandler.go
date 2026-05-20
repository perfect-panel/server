package portal

import (
	"github.com/perfect-panel/server/internal/logic/public/portal"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get available payment methods
func GetAvailablePaymentMethodsHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := portal.NewGetAvailablePaymentMethodsLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetAvailablePaymentMethods()
		result.HttpResult(c, resp, err)
	}
}
