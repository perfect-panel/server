package marketing

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type QueryQuotaTaskListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query quota task list
func NewQueryQuotaTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryQuotaTaskListLogic {
	return &QueryQuotaTaskListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryQuotaTaskListLogic) QueryQuotaTaskList(req *types.QueryQuotaTaskListRequest) (resp *types.QueryQuotaTaskListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
