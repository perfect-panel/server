package payment

import (
	"github.com/perfect-panel/server/internal/logic/admin/payment"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get supported payment platform
func GetPaymentPlatformHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := payment.NewGetPaymentPlatformLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetPaymentPlatform()
		result.HttpResult(c, resp, err)
	}
}
