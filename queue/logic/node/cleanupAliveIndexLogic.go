package node

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

type CleanupAliveIndexLogic struct {
	svc *svc.ServiceContext
}

func NewCleanupAliveIndexLogic(svc *svc.ServiceContext) *CleanupAliveIndexLogic {
	return &CleanupAliveIndexLogic{svc: svc}
}

// ProcessTask scans node:online:uid:index and removes uids whose per-uid ZSet
// has drained to zero. Keeps the index bounded as users churn.
func (l *CleanupAliveIndexLogic) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	if err := l.svc.NodeModel.CleanupOnlineUserUIDIndex(ctx); err != nil {
		logger.Error("[CleanupAliveIndex] cleanup failed", logger.Field("error", err.Error()))
		return err
	}
	return nil
}
