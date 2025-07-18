package handler

import (
	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/svc"
	countrylogic "github.com/perfect-panel/server/queue/logic/country"
	orderLogic "github.com/perfect-panel/server/queue/logic/order"
	smslogic "github.com/perfect-panel/server/queue/logic/sms"
	"github.com/perfect-panel/server/queue/logic/subscription"
	"github.com/perfect-panel/server/queue/logic/traffic"
	"github.com/perfect-panel/server/queue/types"

	emailLogic "github.com/perfect-panel/server/queue/logic/email"
)

func RegisterHandlers(mux *asynq.ServeMux, serverCtx *svc.ServiceContext) {
	// get country task
	mux.Handle(types.ForthwithGetCountry, countrylogic.NewGetNodeCountryLogic(serverCtx))
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
}
