// Package metrics exposes Prometheus counters/gauges for the limiter feature
// (speed + device_limit enforcement). Labels are kept low-cardinality: reason
// is a small enum; per-uid breakdown goes to logs, never to Prometheus.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// LimiterRejectTotal counts Reject events observed by the node limiter.
	// reason is a small enum ("device_limit", future: "expired", etc).
	LimiterRejectTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ppanel_limiter_reject_total",
		Help: "Number of Reject events from the node limiter, partitioned by reason.",
	}, []string{"reason"})

	// AlivelistFetchErrorTotal counts failed alivelist fetches from the node side
	// (reported via node push) OR failed aggregations server-side.
	AlivelistFetchErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ppanel_alivelist_fetch_error_total",
		Help: "Number of failed alivelist aggregations or fetches.",
	})

	// RejectReportErrorTotal counts failed Reject reports (node -> server).
	RejectReportErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ppanel_reject_report_error_total",
		Help: "Number of failed Reject reports from nodes.",
	})

	// AlivelistCacheHitRatio publishes the most recent observed ratio in [0,1].
	// Set atomically by AliveListByUID on each call (not a rolling average).
	AlivelistCacheHitRatio = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ppanel_alivelist_cache_hit_ratio",
		Help: "Latest observed cache hit / total ratio for /alivelist aggregation.",
	})

	// RejectCounterRedisFieldCount tracks the size of user:reject:counter Hash.
	RejectCounterRedisFieldCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ppanel_reject_counter_redis_field_count",
		Help: "Total fields stored in the reject counter Hash (rough scale indicator).",
	})

	// AdminOnlineStatusQPS counts admin online-status lookups (single + batch).
	// Useful to spot leaks/UI hot-loops exposing user IPs.
	AdminOnlineStatusQPS = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ppanel_admin_online_status_qps",
		Help: "Admin online-status endpoint hit count (detail + batch combined).",
	})
)
