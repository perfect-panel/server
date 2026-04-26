package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/metrics"
)

// MaxRejectEventsPerRequest bounds a single /reject payload. Anything above is
// truncated + logged. Protects Redis from a rogue/compromised node spamming events.
const MaxRejectEventsPerRequest = 10000

type ReportRejectLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewReportRejectLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *ReportRejectLogic {
	return &ReportRejectLogic{
		Logger: logger.WithContext(ctx.Request.Context()),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ReportReject ingests aggregated Reject counters from the node limiter.
// Counts are accumulated per (uid, server_id) into RejectCounterKey Hash with
// rolling 24h TTL. Invalid events are skipped without failing the whole batch.
func (l *ReportRejectLogic) ReportReject(req *types.ReportRejectRequest) error {
	events := req.Events
	if len(events) > MaxRejectEventsPerRequest {
		l.Errorw("[ReportReject] payload truncated",
			logger.Field("server_id", req.ServerId),
			logger.Field("received", len(events)),
			logger.Field("kept", MaxRejectEventsPerRequest))
		events = events[:MaxRejectEventsPerRequest]
	}

	for _, e := range events {
		if e.UID <= 0 || e.Count <= 0 {
			continue
		}
		reason := e.Reason
		if reason == "" {
			reason = "unknown"
		}
		metrics.LimiterRejectTotal.WithLabelValues(reason).Add(float64(e.Count))
		if err := l.svcCtx.NodeModel.IncrRejectCount(l.ctx, e.UID, req.ServerId, e.Count, e.Reason); err != nil {
			metrics.RejectReportErrorTotal.Inc()
			l.Errorw("[ReportReject] incr failed",
				logger.Field("uid", e.UID),
				logger.Field("server_id", req.ServerId),
				logger.Field("error", err.Error()))
			// Continue processing remaining events on transient Redis errors.
		}
	}
	return nil
}
