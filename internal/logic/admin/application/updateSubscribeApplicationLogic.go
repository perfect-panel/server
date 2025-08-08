package application

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type UpdateSubscribeApplicationLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update subscribe application
func NewUpdateSubscribeApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSubscribeApplicationLogic {
	return &UpdateSubscribeApplicationLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSubscribeApplicationLogic) UpdateSubscribeApplication(req *types.UpdateSubscribeApplicationRequest) (resp *types.SubscribeApplication, err error) {
	// todo: add your logic here and delete this line

	return
}
