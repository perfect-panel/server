package initialize

import (
	"github.com/perfect-panel/server/internal/svc"
)

func StartInitSystemConfig(svc *svc.ServiceContext) {
	Migrate(svc)
	Site(svc)
	Node(svc)
	Email(svc)
	Device(svc)
	Invite(svc)
	Verify(svc)
	Subscribe(svc)
	Register(svc)
	Mobile(svc)
	if !svc.Config.Debug {
		Telegram(svc)
	}

}
