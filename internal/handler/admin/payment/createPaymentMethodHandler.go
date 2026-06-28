package payment

import (
	"github.com/perfect-panel/server/internal/logic/admin/payment"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Create Payment Method
func CreatePaymentMethodHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.CreatePaymentMethodRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := payment.NewCreatePaymentMethodLogic(c.Request.Context(), svcCtx)
		resp, err := l.CreatePaymentMethod(&req)
		result.HttpResult(c, resp, err)
	}
}
