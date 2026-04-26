package types

const (
	SchedulerCheckSubscription  = "scheduler:check:subscription"
	SchedulerTotalServerData    = "scheduler:total:server"
	SchedulerResetTraffic       = "scheduler:reset:traffic"
	SchedulerTrafficStat        = "scheduler:traffic:stat"
	SchedulerCleanupAliveIndex  = "scheduler:cleanup:alive:index"

	// V4.3 限速→断网状态机 cron(每 5 分钟):
	//   - 90% 阈值预警(决策 20)
	//   - throttled 12h 倒数提醒 / 24h 断网推进(决策 40)
	SchedulerSubscribeTrafficStatus = "scheduler:subscribe:traffic_status"

	// V4.3 audit_log 90 天清理(决策 35)
	SchedulerAuditCleanup = "scheduler:audit:cleanup"

	// V4.3 设备 today_traffic + reset_count_hour/day 凌晨归零
	SchedulerDeviceDailyReset = "scheduler:device:daily_reset"

	// 通知派发:每 60s 拉 notice:queue 渲染并投递(决策 20 / 7.1)
	SchedulerNoticeDispatch = "scheduler:notice:dispatch"

	// V4.3 到期提醒:每天 00:00 扫一次,3 天/1 天预警(决策 7.1)
	SchedulerSubscribeExpireWarn = "scheduler:subscribe:expire_warn"
)
