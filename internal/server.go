package internal

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/perfect-panel/server/initialize"
	"github.com/perfect-panel/server/internal/report"
	"github.com/perfect-panel/server/internal/transport/ginserver"
	"github.com/perfect-panel/server/internal/transport/hertzserver"
	"github.com/perfect-panel/server/pkg/logger"

	"github.com/perfect-panel/server/pkg/proc"
	"github.com/perfect-panel/server/pkg/trace"

	"github.com/perfect-panel/server/internal/svc"
)

type Service struct {
	server transportServer
	svc    *svc.ServiceContext
}

func NewService(svc *svc.ServiceContext) *Service {
	return &Service{
		svc: svc,
	}
}

type transportServer interface {
	Start()
	Shutdown(ctx context.Context) error
}

type ginTransportServer struct {
	server     *http.Server
	tlsEnabled bool
	certFile   string
	keyFile    string
}

func (s *ginTransportServer) Start() {
	var err error
	if s.tlsEnabled {
		err = s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	} else {
		err = s.server.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf("server start error: %s", err.Error())
	}
}

func (s *ginTransportServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func newTransportServer(svc *svc.ServiceContext, addr string) transportServer {
	switch strings.ToLower(svc.Config.Transport.Driver) {
	case "hertz":
		var tlsConfig *tls.Config
		if svc.Config.TLS.Enable {
			cert, err := tls.LoadX509KeyPair(svc.Config.TLS.CertFile, svc.Config.TLS.KeyFile)
			if err != nil {
				logger.Errorf("load tls certificate error: %s", err.Error())
				return nil
			}
			tlsConfig = &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{cert},
			}
		}
		return hertzserver.New(svc, addr, tlsConfig, ginserver.New(svc))
	default:
		return &ginTransportServer{
			server: &http.Server{
				Addr:    addr,
				Handler: ginserver.New(svc),
				TLSConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
			tlsEnabled: svc.Config.TLS.Enable,
			certFile:   svc.Config.TLS.CertFile,
			keyFile:    svc.Config.TLS.KeyFile,
		}
	}
}

func (m *Service) Start() {
	if m.svc == nil {
		panic("config file path is nil")
	}

	// get server port
	port := m.svc.Config.Port
	host := m.svc.Config.Host
	// check gateway mode
	if report.IsGatewayMode() {
		// get free port
		freePort, err := report.ModulePort()
		if err != nil {
			logger.Errorf("get module port error: %s", err.Error())
			panic(err)
		}
		port = freePort
		host = "127.0.0.1"
		// register module
		err = report.RegisterModule(port)
		if err != nil {
			logger.Errorf("register module error: %s", err.Error())
			os.Exit(1)
		}
		logger.Infof("module registered on port %d", port)
	}

	serverAddr := fmt.Sprintf("%v:%d", host, port)
	initialize.StartInitSystemConfig(m.svc)
	m.server = newTransportServer(m.svc, serverAddr)
	if m.server == nil {
		return
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
	m.server.Start()
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
