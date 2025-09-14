package marketing

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/task"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	types2 "github.com/perfect-panel/server/queue/types"
	"gorm.io/gorm"
)

type CreateBatchSendEmailTaskLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateBatchSendEmailTaskLogic Create a batch send email task
func NewCreateBatchSendEmailTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateBatchSendEmailTaskLogic {
	return &CreateBatchSendEmailTaskLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
func (l *CreateBatchSendEmailTaskLogic) CreateBatchSendEmailTask(req *types.CreateBatchSendEmailTaskRequest) (err error) {
	tx := l.svcCtx.DB

	var emails []string

	// 通用查询器（含 user JOIN + 注册时间范围过滤）
	baseQuery := func() *gorm.DB {
		query := tx.Model(&user.AuthMethods{}).
			Select("auth_identifier").
			Joins("JOIN user ON user.id = user_auth_methods.user_id").
			Where("auth_type = ?", "email")

		if req.RegisterStartTime != 0 {
			query = query.Where("user.created_at >= ?", time.UnixMilli(req.RegisterStartTime))
		}
		if req.RegisterEndTime != 0 {
			query = query.Where("user.created_at <= ?", time.UnixMilli(req.RegisterEndTime))
		}
		return query
	}

	var query *gorm.DB

	scope := task.ParseScopeType(req.Scope)

	switch scope {
	case task.ScopeAll:
		query = baseQuery()

	case task.ScopeActive:
		query = baseQuery().
			Joins("JOIN user_subscribe ON user.id = user_subscribe.user_id").
			Where("user_subscribe.status IN ?", []int64{1, 2})

	case task.ScopeExpired:
		query = baseQuery().
			Joins("JOIN user_subscribe ON user.id = user_subscribe.user_id").
			Where("user_subscribe.status = ?", 3)

	case task.ScopeNone:
		query = baseQuery().
			Joins("LEFT JOIN user_subscribe ON user.id = user_subscribe.user_id").
			Where("user_subscribe.user_id IS NULL")
	default:

	}
	if query != nil {
		// 执行查询
		err = query.Pluck("auth_identifier", &emails).Error
		if err != nil {
			l.Errorf("[CreateBatchSendEmailTask] Failed to fetch email addresses: %v", err.Error())
			return xerr.NewErrCode(xerr.DatabaseQueryError)
		}
	}

	// 邮箱列表为空，返回错误
	if len(emails) == 0 && scope != task.ScopeSkip {
		l.Errorf("[CreateBatchSendEmailTask] No email addresses found for the specified scope")
		return xerr.NewErrMsg("No email addresses found for the specified scope")
	}

	// 邮箱地址去重
	emails = tool.RemoveDuplicateElements(emails...)

	var additionalEmails []string
	// 追加额外的邮箱地址（不覆盖）
	if req.Additional != "" {
		additionalEmails = tool.RemoveDuplicateElements(strings.Split(req.Additional, "\n")...)
	}
	if len(additionalEmails) == 0 && scope == task.ScopeSkip {
		l.Errorf("[CreateBatchSendEmailTask] No additional email addresses provided for skip scope")
		return xerr.NewErrMsg("No additional email addresses provided for skip scope")
	}

	scheduledAt := time.Now().Add(10 * time.Second) // 默认延迟10秒执行,防止任务创建和执行时间过于接近
	if req.Scheduled != 0 {
		scheduledAt = time.Unix(req.Scheduled, 0)
		if scheduledAt.Before(time.Now()) {
			scheduledAt = time.Now()
		}
	}

	scopeInfo := task.EmailScope{
		Type:              scope.Int8(),
		RegisterStartTime: req.RegisterStartTime,
		RegisterEndTime:   req.RegisterEndTime,
		Recipients:        emails,
		Additional:        additionalEmails,
		Scheduled:         req.Scheduled,
		Interval:          req.Interval,
		Limit:             req.Limit,
	}
	scopeBytes, _ := scopeInfo.Marshal()

	taskContent := task.EmailContent{
		Subject: req.Subject,
		Content: req.Content,
	}

	contentBytes, _ := taskContent.Marshal()

	var total uint64
	if additionalEmails != nil {
		list := append(emails, additionalEmails...)
		total = uint64(len(tool.RemoveDuplicateElements(list...)))
	} else {
		total = uint64(len(emails))
	}

	taskInfo := &task.Task{
		Type:    task.TypeEmail,
		Scope:   string(scopeBytes),
		Content: string(contentBytes),
		Status:  0,
		Errors:  "",
		Total:   total,
		Current: 0,
	}

	if err = l.svcCtx.DB.Model(&task.Task{}).Create(taskInfo).Error; err != nil {
		l.Errorf("[CreateBatchSendEmailTask] Failed to create email task: %v", err.Error())
		return xerr.NewErrCode(xerr.DatabaseInsertError)
	}
	// create task
	l.Infof("[CreateBatchSendEmailTask] Successfully created email task with ID: %d", taskInfo.Id)

	t := asynq.NewTask(types2.ScheduledBatchSendEmail, []byte(strconv.FormatInt(taskInfo.Id, 10)))
	info, err := l.svcCtx.Queue.EnqueueContext(l.ctx, t, asynq.ProcessAt(scheduledAt))
	if err != nil {
		l.Errorf("[CreateBatchSendEmailTask] Failed to enqueue email task: %v", err.Error())
		return xerr.NewErrCode(xerr.QueueEnqueueError)
	}
	l.Infof("[CreateBatchSendEmailTask] Successfully enqueued email task with ID: %s, scheduled at: %s", info.ID, scheduledAt.Format(time.DateTime))

	return nil
}
