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
	Currency(svc)
	// V4.3 决策 25:确保 11 款官方客户端 + 教程占位存在,删除遗留 Default。
	Application(svc)
	// V4.3:把老订阅 device_count < plan.device_limit 的补齐(一次性纠正)。
	SubscribeDeviceBackfill(svc)
	if !svc.Config.Debug {
		Telegram(svc)
	}

}
