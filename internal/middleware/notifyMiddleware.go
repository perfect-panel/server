package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/constant"
)

type PaymentParams struct {
	Platform string `uri:"platform"`
	Token    string `uri:"token"`
}

func NotifyMiddleware(svc *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params PaymentParams
		// Get platform and token from uri
		if err := c.ShouldBindUri(&params); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		ctx, err := PaymentNotifyContext(c.Request.Context(), svc, params.Token)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func PaymentNotifyContext(ctx context.Context, svc *svc.ServiceContext, token string) (context.Context, error) {
	config, err := svc.Store.Payment().FindOneByPaymentToken(ctx, token)
	if err != nil {
		return ctx, err
	}
	ctx = context.WithValue(ctx, constant.CtxKeyPlatform, config.Platform)
	ctx = context.WithValue(ctx, constant.CtxKeyPayment, config)
	return ctx, nil
}
