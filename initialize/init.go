package initialize

import (
	"github.com/perfect-panel/ppanel-server/internal/svc"
)

func StartInitSystemConfig(svc *svc.ServiceContext) {
	Migrate(svc)
	Site(svc)
	Node(svc)
	Email(svc)
	Invite(svc)
	Verify(svc)
	Subscribe(svc)
	Register(svc)
	Mobile(svc)
	TrafficDataToRedis(svc)
	if !svc.Config.Debug {
		Telegram(svc)
	}

}
