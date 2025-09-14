package emailLogic

import (
	"context"
	"strconv"

	"github.com/hibiken/asynq"
	taskModel "github.com/perfect-panel/server/internal/model/task"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/email"
	"github.com/perfect-panel/server/pkg/logger"
)

type BatchEmailLogic struct {
	svcCtx *svc.ServiceContext
}

type ErrorInfo struct {
	Error string `json:"error"`
	Email string `json:"email"`
	Time  int64  `json:"time"`
}

func NewBatchEmailLogic(svcCtx *svc.ServiceContext) *BatchEmailLogic {
	return &BatchEmailLogic{
		svcCtx: svcCtx,
	}
}

func (l *BatchEmailLogic) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// 解析任务负载
	payload := task.Payload()
	if len(payload) == 0 {
		logger.Error("[BatchEmailLogic] ProcessTask failed: empty payload")
		return asynq.SkipRetry
	}
	// 转换获取任务id
	taskID, err := strconv.ParseInt(string(payload), 10, 64)
	if err != nil {
		logger.WithContext(ctx).Error("[BatchEmailLogic] ProcessTask failed: invalid task ID",
			logger.Field("error", err.Error()),
			logger.Field("payload", string(payload)),
		)
		return asynq.SkipRetry
	}
	tx := l.svcCtx.DB.WithContext(ctx)
	var taskInfo taskModel.Task
	if err = tx.Model(&taskModel.Task{}).Where("id = ?", taskID).First(&taskInfo).Error; err != nil {
		logger.WithContext(ctx).Error("[BatchEmailLogic] ProcessTask failed",
			logger.Field("error", err.Error()),
			logger.Field("taskID", taskID),
		)
		return asynq.SkipRetry
	}

	if taskInfo.Status != 0 {
		logger.WithContext(ctx).Info("[BatchEmailLogic] ProcessTask skipped: task already processed",
			logger.Field("taskID", taskID),
			logger.Field("status", taskInfo.Status),
		)
		return nil
	}

	sender, err := email.NewSender(l.svcCtx.Config.Email.Platform, l.svcCtx.Config.Email.PlatformConfig, l.svcCtx.Config.Site.SiteName)
	if err != nil {
		logger.WithContext(ctx).Error("[BatchEmailLogic] NewSender failed", logger.Field("error", err.Error()))
		return nil
	}
	manager := email.NewWorkerManager(l.svcCtx.DB, sender)
	if manager == nil {
		logger.WithContext(ctx).Error("[BatchEmailLogic] ProcessTask failed: worker manager is nil")
		return asynq.SkipRetry
	}

	// 添加或获取 Worker 实例
	manager.AddWorker(taskID)
	return nil
}
