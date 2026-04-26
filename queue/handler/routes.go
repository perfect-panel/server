package handler

import (
	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/svc"
	nodeLogic "github.com/perfect-panel/server/queue/logic/node"
	orderLogic "github.com/perfect-panel/server/queue/logic/order"
	smslogic "github.com/perfect-panel/server/queue/logic/sms"
	"github.com/perfect-panel/server/queue/logic/subscription"
	"github.com/perfect-panel/server/queue/logic/task"
	"github.com/perfect-panel/server/queue/logic/traffic"
	"github.com/perfect-panel/server/queue/types"

	emailLogic "github.com/perfect-panel/server/queue/logic/email"
)

func RegisterHandlers(mux *asynq.ServeMux, serverCtx *svc.ServiceContext) {
	// Send email task
	mux.Handle(types.ForthwithSendEmail, emailLogic.NewSendEmailLogic(serverCtx))
	// Send sms task
	mux.Handle(types.ForthwithSendSms, smslogic.NewSendSmsLogic(serverCtx))
	// Defer close order task
	mux.Handle(types.DeferCloseOrder, orderLogic.NewDeferCloseOrderLogic(serverCtx))
	// Forthwith activate order task
	mux.Handle(types.ForthwithActivateOrder, orderLogic.NewActivateOrderLogic(serverCtx))

	// Forthwith traffic statistics
	mux.Handle(types.ForthwithTrafficStatistics, traffic.NewTrafficStatisticsLogic(serverCtx))

	// Schedule check subscription
	mux.Handle(types.SchedulerCheckSubscription, subscription.NewCheckSubscriptionLogic(serverCtx))

	// Schedule total server data
	mux.Handle(types.SchedulerTotalServerData, traffic.NewServerDataLogic(serverCtx))

	// Schedule reset traffic
	mux.Handle(types.SchedulerResetTraffic, traffic.NewResetTrafficLogic(serverCtx))

	// ScheduledBatchSendEmail
	mux.Handle(types.ScheduledBatchSendEmail, emailLogic.NewBatchEmailLogic(serverCtx))

	// ScheduledTrafficStat
	mux.Handle(types.SchedulerTrafficStat, traffic.NewStatLogic(serverCtx))

	// ForthwithQuotaTask
	mux.Handle(types.ForthwithQuotaTask, task.NewQuotaTaskLogic(serverCtx))

	// Schedule cleanup alive index (per-uid online ZSet orphan uids)
	mux.Handle(types.SchedulerCleanupAliveIndex, nodeLogic.NewCleanupAliveIndexLogic(serverCtx))

	// V4.3 限速→断网状态机:每 5 分钟推进 90% / 12h / 24h
	mux.Handle(types.SchedulerSubscribeTrafficStatus, subscription.NewTrafficStatusLogic(serverCtx))

	// V4.3 audit_log 90 天清理
	mux.Handle(types.SchedulerAuditCleanup, subscription.NewAuditCleanupLogic(serverCtx))

	// V4.3 设备 today_traffic + reset_count 凌晨归零
	mux.Handle(types.SchedulerDeviceDailyReset, subscription.NewDeviceDailyResetLogic(serverCtx))

	// V4.3 通知派发(渲染 site_content 模板 → email/TG)
	mux.Handle(types.SchedulerNoticeDispatch, subscription.NewNoticeDispatchLogic(serverCtx))

	// V4.3 到期提醒(每天 00:00,3d/1d 预警)
	mux.Handle(types.SchedulerSubscribeExpireWarn, subscription.NewExpireWarnLogic(serverCtx))
}
