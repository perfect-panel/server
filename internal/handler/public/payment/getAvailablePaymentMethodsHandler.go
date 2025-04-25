package payment

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/public/payment"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Get available payment methods
func GetAvailablePaymentMethodsHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := payment.NewGetAvailablePaymentMethodsLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetAvailablePaymentMethods()
		result.HttpResult(c, resp, err)
	}
}
