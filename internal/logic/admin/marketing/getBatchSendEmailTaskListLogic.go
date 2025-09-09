package marketing

import (
	"context"
	"strings"

	"github.com/perfect-panel/server/internal/model/task"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
)

type GetBatchSendEmailTaskListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetBatchSendEmailTaskListLogic Get batch send email task list
func NewGetBatchSendEmailTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBatchSendEmailTaskListLogic {
	return &GetBatchSendEmailTaskListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBatchSendEmailTaskListLogic) GetBatchSendEmailTaskList(req *types.GetBatchSendEmailTaskListRequest) (resp *types.GetBatchSendEmailTaskListResponse, err error) {

	var tasks []*task.Task
	tx := l.svcCtx.DB.Model(&task.Task{}).Where("`type` = ?", task.TypeEmail)
	if req.Status != nil {
		tx = tx.Where("status = ?", *req.Status)
	}
	if req.Scope != nil {
		tx = tx.Where("scope = ?", req.Scope)
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 10
	}
	err = tx.Offset((req.Page - 1) * req.Size).Limit(req.Size).Order("created_at DESC").Find(&tasks).Error
	if err != nil {
		l.Errorf("failed to get email tasks: %v", err)
		return nil, xerr.NewErrCode(xerr.DatabaseQueryError)
	}

	list := make([]types.BatchSendEmailTask, 0)

	for _, t := range tasks {
		var scopeInfo task.EmailScope
		if err = scopeInfo.Unmarshal([]byte(t.Scope)); err != nil {
			l.Errorf("[GetBatchSendEmailTaskList] failed to unmarshal email task scope: %v", err.Error())
			continue
		}
		var contentInfo task.EmailContent
		if err = contentInfo.Unmarshal([]byte(t.Content)); err != nil {
			l.Errorf("[GetBatchSendEmailTaskList] failed to unmarshal email task content: %v", err.Error())
			continue
		}

		list = append(list, types.BatchSendEmailTask{
			Id:                t.Id,
			Subject:           contentInfo.Subject,
			Content:           contentInfo.Content,
			Recipients:        strings.Join(scopeInfo.Recipients, "\n"),
			Scope:             scopeInfo.Type,
			RegisterStartTime: scopeInfo.RegisterStartTime,
			RegisterEndTime:   scopeInfo.RegisterEndTime,
			Additional:        strings.Join(scopeInfo.Additional, "\n"),
			Scheduled:         scopeInfo.Scheduled,
			Interval:          scopeInfo.Interval,
			Limit:             scopeInfo.Limit,
			Status:            uint8(t.Status),
			Errors:            t.Errors,
			Total:             t.Total,
			Current:           t.Current,
			CreatedAt:         t.CreatedAt.UnixMilli(),
			UpdatedAt:         t.UpdatedAt.UnixMilli(),
		})
	}

	return &types.GetBatchSendEmailTaskListResponse{
		List: list,
	}, nil
}
