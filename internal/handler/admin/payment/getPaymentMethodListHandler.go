package payment

import (
	"github.com/perfect-panel/server/internal/logic/admin/payment"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// GetPaymentMethodListHandler Get Payment Method List
func GetPaymentMethodListHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.GetPaymentMethodListRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := payment.NewGetPaymentMethodListLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetPaymentMethodList(&req)
		result.HttpResult(c, resp, err)
	}
}
