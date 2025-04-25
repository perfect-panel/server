package internal

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/perfect-panel/server/pkg/logger"

	"github.com/perfect-panel/server/pkg/proc"
	"github.com/perfect-panel/server/pkg/trace"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/initialize"
	"github.com/perfect-panel/server/internal/handler"
	"github.com/perfect-panel/server/internal/middleware"
	"github.com/perfect-panel/server/internal/svc"
)

type Service struct {
	server *http.Server
	svc    *svc.ServiceContext
}

func NewService(svc *svc.ServiceContext) *Service {
	return &Service{
		svc: svc,
	}
}

func initServer(svc *svc.ServiceContext) *gin.Engine {

	// start init system config
	initialize.StartInitSystemConfig(svc)
	// init gin server
	r := gin.Default()
	r.RemoteIPHeaders = []string{"X-Original-Forwarded-For", "X-Forwarded-For", "X-Real-IP"}
	// init session
	sessionStore, err := redis.NewStore(10, "tcp", svc.Config.Redis.Host, svc.Config.Redis.Pass, []byte(svc.Config.JwtAuth.AccessSecret))
	if err != nil {
		logger.Errorw("init session error", logger.Field("error", err.Error()))
		panic(err)
	}
	r.Use(sessions.Sessions("ppanel", sessionStore))
	// use cors middleware
	r.Use(middleware.TraceMiddleware(svc), middleware.LoggerMiddleware(svc), middleware.CorsMiddleware, middleware.PanDomainMiddleware(svc), gin.Recovery())

	// register handlers
	handler.RegisterHandlers(r, svc)
	// register subscribe handler
	handler.RegisterSubscribeHandlers(r, svc)
	// register telegram handler
	handler.RegisterTelegramHandlers(r, svc)
	// register notify handler
	handler.RegisterNotifyHandlers(r, svc)
	return r
}

func (m *Service) Start() {
	if m.svc == nil {
		panic("config file path is nil")
	}
	// init service
	r := initServer(m.svc)
	serverAddr := fmt.Sprintf("%v:%d", m.svc.Config.Host, m.svc.Config.Port)
	m.server = &http.Server{
		Addr:    serverAddr,
		Handler: r,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	trace.StartAgent(trace.Config{
		Name:    "ppanel",
		Sampler: 1.0,
		Batcher: "",
	})
	proc.AddShutdownListener(func() {
		trace.StopAgent()
	})
	m.svc.Restart = m.Restart
	logger.Infof("server start at %v", serverAddr)
	if m.svc.Config.TLS.Enable {
		if err := m.server.ListenAndServeTLS(m.svc.Config.TLS.CertFile, m.svc.Config.TLS.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("server start error: %s", err.Error())
		}
	} else {
		if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("server start error: %s", err.Error())
		}
	}
}

func (m *Service) Stop() {
	if m.server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.server.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown error: %s", err.Error())
	}
	logger.Info("server shutdown")
}

func (m *Service) Restart() error {
	if m.server == nil {
		return errors.New("server is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.server.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown error: %v", err.Error())
		return err
	}
	logger.Info("server shutdown")
	go m.Start()
	return nil
}
