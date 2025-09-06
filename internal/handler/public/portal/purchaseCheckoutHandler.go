package portal

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/portal"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// PurchaseCheckoutHandler Purchase Checkout
func PurchaseCheckoutHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.CheckoutOrderRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := portal.NewPurchaseCheckoutLogic(c.Request.Context(), svcCtx)
		resp, err := l.PurchaseCheckout(&req)
		result.HttpResult(c, resp, err)
	}
}
