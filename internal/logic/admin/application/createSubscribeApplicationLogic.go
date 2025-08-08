package application

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type CreateSubscribeApplicationLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create subscribe application
func NewCreateSubscribeApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSubscribeApplicationLogic {
	return &CreateSubscribeApplicationLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSubscribeApplicationLogic) CreateSubscribeApplication(req *types.CreateSubscribeApplicationRequest) (resp *types.SubscribeApplication, err error) {
	// todo: add your logic here and delete this line

	return
}
