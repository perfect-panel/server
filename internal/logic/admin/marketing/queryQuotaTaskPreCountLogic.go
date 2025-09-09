package marketing

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type QueryQuotaTaskPreCountLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Query quota task pre-count
func NewQueryQuotaTaskPreCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryQuotaTaskPreCountLogic {
	return &QueryQuotaTaskPreCountLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryQuotaTaskPreCountLogic) QueryQuotaTaskPreCount(req *types.QueryQuotaTaskPreCountRequest) (resp *types.QueryQuotaTaskPreCountResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
