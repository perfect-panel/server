package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/handler/notify"
	"github.com/perfect-panel/ppanel-server/internal/middleware"
	"github.com/perfect-panel/ppanel-server/internal/svc"
)

func RegisterNotifyHandlers(router *gin.Engine, serverCtx *svc.ServiceContext) {
	group := router.Group("/v1/notify/")
	group.Use(middleware.NotifyMiddleware(serverCtx))
	{
		group.Any("/:platform/:token", notify.PaymentNotifyHandler(serverCtx))
	}

}
