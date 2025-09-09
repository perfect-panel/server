package marketing

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type QueryQuotaTaskStatusLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query quota task status
func NewQueryQuotaTaskStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryQuotaTaskStatusLogic {
	return &QueryQuotaTaskStatusLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryQuotaTaskStatusLogic) QueryQuotaTaskStatus(req *types.QueryQuotaTaskStatusRequest) (resp *types.QueryQuotaTaskStatusResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
