package middleware

import (
	"context"

	"github.com/perfect-panel/server/pkg/constant"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/svc"
)

type PaymentParams struct {
	Platform string `uri:"platform"`
	Token    string `uri:"token"`
}

func NotifyMiddleware(svc *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var params PaymentParams
		// Get platform and token from uri
		if err := c.ShouldBindUri(&params); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		config, err := svc.PaymentModel.FindOneByPaymentToken(ctx, params.Token)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		ctx = context.WithValue(ctx, constant.CtxKeyPlatform, config.Platform)
		ctx = context.WithValue(ctx, constant.CtxKeyPayment, config)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
