package application

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type GetSubscribeApplicationListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get subscribe application list
func NewGetSubscribeApplicationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubscribeApplicationListLogic {
	return &GetSubscribeApplicationListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubscribeApplicationListLogic) GetSubscribeApplicationList(req *types.GetSubscribeApplicationListRequest) (resp *types.GetSubscribeApplicationListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
