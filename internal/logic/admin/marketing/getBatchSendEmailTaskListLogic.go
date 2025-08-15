package marketing

import (
	"context"

	"github.com/perfect-panel/server/internal/model/task"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
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

	var tasks []*task.EmailTask
	tx := l.svcCtx.DB.Model(&task.EmailTask{})
	if req.Status != nil {
		tx = tx.Where("status = ?", *req.Status)
	}
	if req.Scope != "" {
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
	tool.DeepCopy(&list, tasks)
	return &types.GetBatchSendEmailTaskListResponse{
		List: list,
	}, nil
}
