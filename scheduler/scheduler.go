package scheduler

import (
	"time"

	"github.com/perfect-panel/server/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/queue/types"
)

type Service struct {
	svc    *svc.ServiceContext
	server *asynq.Scheduler
}

func NewService(svc *svc.ServiceContext) *Service {
	return &Service{
		svc:    svc,
		server: initService(svc),
	}
}

func (m *Service) Start() {
	logger.Infof("start scheduler service")
	// schedule check subscription task: every 60 seconds
	checkTask := asynq.NewTask(types.SchedulerCheckSubscription, nil)
	if _, err := m.server.Register("@every 60s", checkTask); err != nil {
		logger.Errorf("register check subscription task failed: %s", err.Error())
	}
	// schedule total server data task: every 5 minutes
	totalServerDataTask := asynq.NewTask(types.SchedulerTotalServerData, nil)
	if _, err := m.server.Register("@every 180s", totalServerDataTask); err != nil {
		logger.Errorf("register total server data task failed: %s", err.Error())
	}
	// schedule reset traffic task: every 24 hours
	resetTrafficTask := asynq.NewTask(types.SchedulerResetTraffic, nil)
	if _, err := m.server.Register("@every 24h", resetTrafficTask); err != nil {
		logger.Errorf("register reset traffic task failed: %s", err.Error())
	}

	if err := m.server.Run(); err != nil {
		logger.Errorf("run scheduler failed: %s", err.Error())
	}
}

func (m *Service) Stop() {
	logger.Info("stop scheduler service")
	m.server.Shutdown()
}

func initService(svc *svc.ServiceContext) *asynq.Scheduler {
	location, _ := time.LoadLocation("Asia/Shanghai")
	return asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: svc.Config.Redis.Host, Password: svc.Config.Redis.Pass, DB: 5},
		&asynq.SchedulerOpts{
			Location: location,
		},
	)
}
