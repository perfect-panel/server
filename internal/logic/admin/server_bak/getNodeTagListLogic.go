package server_bak

import (
	"context"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
)

type GetNodeTagListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get node tag list
func NewGetNodeTagListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodeTagListLogic {
	return &GetNodeTagListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNodeTagListLogic) GetNodeTagList() (resp *types.GetNodeTagListResponse, err error) {
	tags, err := l.svcCtx.ServerModel.FindServerTags(l.ctx)
	return &types.GetNodeTagListResponse{
		Tags: tool.RemoveDuplicateElements(tags...),
	}, nil
}
