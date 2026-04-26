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
	//// schedule total server data task: every 5 minutes
	//totalServerDataTask := asynq.NewTask(types.SchedulerTotalServerData, nil)
	//if _, err := m.server.Register("@every 180s", totalServerDataTask); err != nil {
	//	logger.Errorf("register total server data task failed: %s", err.Error())
	//}
	// schedule reset traffic task: every day at 00:30
	resetTrafficTask := asynq.NewTask(types.SchedulerResetTraffic, nil)
	if _, err := m.server.Register("30 0 * * *", resetTrafficTask); err != nil {
		logger.Errorf("register reset traffic task failed: %s", err.Error())
	}

	// schedule traffic stat task: every day at 00:00
	trafficStatTask := asynq.NewTask(types.SchedulerTrafficStat, nil)
	if _, err := m.server.Register("0 0 * * *", trafficStatTask, asynq.MaxRetry(3)); err != nil {
		logger.Errorf("register traffic stat task failed: %s", err.Error())
	}

	// schedule update exchange rate task: every day at 01:00
	rateTask := asynq.NewTask(types.ForthwithQuotaTask, nil)
	if _, err := m.server.Register("0 1 * * *", rateTask, asynq.MaxRetry(3)); err != nil {
		logger.Errorf("register update exchange rate task failed: %s", err.Error())
	}

	// schedule cleanup alive uid index: every 5 minutes
	cleanupAliveTask := asynq.NewTask(types.SchedulerCleanupAliveIndex, nil)
	if _, err := m.server.Register("@every 300s", cleanupAliveTask); err != nil {
		logger.Errorf("register cleanup alive index task failed: %s", err.Error())
	}

	// V4.3 限速→断网状态机:每 5 分钟推进 90% / 12h / 24h
	trafficStatusTask := asynq.NewTask(types.SchedulerSubscribeTrafficStatus, nil)
	if _, err := m.server.Register("@every 300s", trafficStatusTask); err != nil {
		logger.Errorf("register subscribe traffic status task failed: %s", err.Error())
	}

	// V4.3 audit_log 90 天清理:每天 03:00
	auditCleanupTask := asynq.NewTask(types.SchedulerAuditCleanup, nil)
	if _, err := m.server.Register("0 3 * * *", auditCleanupTask, asynq.MaxRetry(2)); err != nil {
		logger.Errorf("register audit cleanup task failed: %s", err.Error())
	}

	// V4.3 设备 today_traffic + reset_count 凌晨归零:每天 00:05(避开 reset traffic 0:30)
	deviceDailyTask := asynq.NewTask(types.SchedulerDeviceDailyReset, nil)
	if _, err := m.server.Register("5 0 * * *", deviceDailyTask, asynq.MaxRetry(2)); err != nil {
		logger.Errorf("register device daily reset task failed: %s", err.Error())
	}

	// V4.3 通知派发:每 60s 拉 Redis list 渲染并投递
	noticeDispatchTask := asynq.NewTask(types.SchedulerNoticeDispatch, nil)
	if _, err := m.server.Register("@every 60s", noticeDispatchTask); err != nil {
		logger.Errorf("register notice dispatch task failed: %s", err.Error())
	}

	// V4.3 到期提醒:每天 00:00 扫一次,3 天/1 天预警
	expireWarnTask := asynq.NewTask(types.SchedulerSubscribeExpireWarn, nil)
	if _, err := m.server.Register("0 0 * * *", expireWarnTask, asynq.MaxRetry(2)); err != nil {
		logger.Errorf("register subscribe expire warn task failed: %s", err.Error())
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
