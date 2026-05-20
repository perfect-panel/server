package handler

import (
	"github.com/perfect-panel/server/internal/handler/notify"
	"github.com/perfect-panel/server/internal/middleware"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
)

func RegisterNotifyHandlers(router *hertzx.Engine, serverCtx *svc.ServiceContext) {
	group := router.Group("/v1/notify/")
	group.Use(middleware.NotifyMiddleware(serverCtx))
	{
		group.Any("/:platform/:token", notify.PaymentNotifyHandler(serverCtx))
	}

}
