package marketing

import (
	"context"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/task"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	queueType "github.com/perfect-panel/server/queue/types"
	"github.com/pkg/errors"
)

type CreateQuotaTaskLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateQuotaTaskLogic Create a quota task
func NewCreateQuotaTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateQuotaTaskLogic {
	return &CreateQuotaTaskLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateQuotaTaskLogic) CreateQuotaTask(req *types.CreateQuotaTaskRequest) error {
	var subs []*user.Subscribe
	query := l.svcCtx.DB.WithContext(l.ctx).Model(&user.Subscribe{})
	if len(req.Subscribers) > 0 {
		query = query.Where("`subscribe_id` IN ?", req.Subscribers)
	}

	if req.IsActive != nil && *req.IsActive {
		query = query.Where("`status` IN ?", []int64{0, 1, 2}) // 0: Pending 1: Active 2: Finished
	}
	if req.StartTime != 0 {
		start := time.UnixMilli(req.StartTime)
		query = query.Where("`start_time` <= ?", start)
	}
	if req.EndTime != 0 {
		end := time.UnixMilli(req.EndTime)
		query = query.Where("`expire_time` >= ?", end)
	}

	if err := query.Find(&subs).Error; err != nil {
		l.Errorf("[CreateQuotaTask] find subscribers error: %v", err.Error())
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribers error")
	}
	if len(subs) == 0 {
		return errors.Wrapf(xerr.NewErrMsg("No subscribers found"), "no subscribers found")
	}
	var subIds []int64
	for _, sub := range subs {
		subIds = append(subIds, sub.Id)
	}

	scopeInfo := task.QuotaScope{
		Subscribers: req.Subscribers,
		IsActive:    req.IsActive,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Objects:     subIds,
	}
	scopeBytes, _ := scopeInfo.Marshal()
	contentInfo := task.QuotaContent{
		ResetTraffic: req.ResetTraffic,
		Days:         req.Days,
		GiftType:     req.GiftType,
		GiftValue:    req.GiftValue,
	}
	contentBytes, _ := contentInfo.Marshal()
	// create task
	newTask := &task.Task{
		Type:    task.TypeQuota,
		Status:  0,
		Scope:   string(scopeBytes),
		Content: string(contentBytes),
		Total:   uint64(len(subIds)),
		Current: 0,
		Errors:  "",
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&task.Task{}).Create(newTask).Error; err != nil {
		l.Errorf("[CreateQuotaTask] create task error: %v", err.Error())
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create task error")
	}

	// enqueue task
	queueTask := asynq.NewTask(queueType.ForthwithQuotaTask, []byte(strconv.FormatInt(newTask.Id, 10)))
	if _, err := l.svcCtx.Queue.EnqueueContext(l.ctx, queueTask); err != nil {
		l.Errorf("[CreateQuotaTask] enqueue task error: %v", err.Error())
		return errors.Wrapf(xerr.NewErrCode(xerr.QueueEnqueueError), "enqueue task error")
	}
	logger.Infof("[CreateQuotaTask] Successfully created task with ID: %d", newTask.Id)
	return nil
}
