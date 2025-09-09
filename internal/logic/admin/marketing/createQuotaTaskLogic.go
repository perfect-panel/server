package marketing

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type CreateQuotaTaskLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create a quota task
func NewCreateQuotaTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateQuotaTaskLogic {
	return &CreateQuotaTaskLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateQuotaTaskLogic) CreateQuotaTask(req *types.CreateQuotaTaskRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
