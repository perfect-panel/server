package marketing

import (
	"context"

	"github.com/perfect-panel/server/internal/model/task"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryQuotaTaskStatusLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewQueryQuotaTaskStatusLogic Query quota task status
func NewQueryQuotaTaskStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryQuotaTaskStatusLogic {
	return &QueryQuotaTaskStatusLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryQuotaTaskStatusLogic) QueryQuotaTaskStatus(req *types.QueryQuotaTaskStatusRequest) (resp *types.QueryQuotaTaskStatusResponse, err error) {
	var data *task.Task
	err = l.svcCtx.DB.Model(&task.Task{}).Where("id = ? AND `type` = ?", req.Id, task.TypeQuota).First(&data).Error
	if err != nil {
		l.Errorf("[QueryQuotaTaskStatus] failed to get quota task: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), " failed to get quota task: %v", err.Error())
	}
	return &types.QueryQuotaTaskStatusResponse{
		Status:  uint8(data.Status),
		Current: int64(data.Current),
		Total:   int64(data.Total),
		Errors:  data.Errors,
	}, nil
}
