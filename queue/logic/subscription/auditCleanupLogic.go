package subscription

// V4.3 audit_log 90 天保留(决策 35)。每天 03:00 跑一次。

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

const auditRetentionDays = 90

type AuditCleanupLogic struct {
	svc *svc.ServiceContext
}

func NewAuditCleanupLogic(svc *svc.ServiceContext) *AuditCleanupLogic {
	return &AuditCleanupLogic{svc: svc}
}

func (l *AuditCleanupLogic) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	cutoff := time.Now().AddDate(0, 0, -auditRetentionDays)
	rows, err := l.svc.AuditModel.CleanOlderThan(ctx, cutoff)
	if err != nil {
		logger.WithContext(ctx).Error("[AuditCleanup] failed",
			logger.Field("error", err.Error()), logger.Field("cutoff", cutoff))
		return nil
	}
	if rows > 0 {
		logger.WithContext(ctx).Info("[AuditCleanup] purged",
			logger.Field("rows", rows), logger.Field("cutoff", cutoff))
	}
	return nil
}
