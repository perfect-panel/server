package server

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type ResetSortWithNodeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Reset node sort
func NewResetSortWithNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetSortWithNodeLogic {
	return &ResetSortWithNodeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetSortWithNodeLogic) ResetSortWithNode(req *types.ResetSortRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
