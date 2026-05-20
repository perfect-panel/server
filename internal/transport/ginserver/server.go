package ginserver

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/initialize"
	"github.com/perfect-panel/server/internal/handler"
	"github.com/perfect-panel/server/internal/middleware"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

func New(svc *svc.ServiceContext) *gin.Engine {
	initialize.StartInitSystemConfig(svc)

	r := gin.Default()
	r.RemoteIPHeaders = []string{"X-Original-Forwarded-For", "X-Forwarded-For", "X-Real-IP"}

	sessionStore, err := redis.NewStore(10, "tcp", svc.Config.Redis.Host, svc.Config.Redis.Pass, []byte(svc.Config.JwtAuth.AccessSecret))
	if err != nil {
		logger.Errorw("init session error", logger.Field("error", err.Error()))
		panic(err)
	}
	r.Use(sessions.Sessions("ppanel", sessionStore))
	r.Use(middleware.TraceMiddleware(svc), middleware.LoggerMiddleware(svc), middleware.CorsMiddleware, gin.Recovery())

	handler.RegisterHandlers(r, svc)
	handler.RegisterSubscribeHandlers(r, svc)
	handler.RegisterTelegramHandlers(r, svc)
	handler.RegisterNotifyHandlers(r, svc)
	return r
}
