package server

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetNodeGroupListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNodeGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodeGroupListLogic {
	return &GetNodeGroupListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNodeGroupListLogic) GetNodeGroupList() (resp *types.GetNodeGroupListResponse, err error) {
	nodeGroupList, err := l.svcCtx.ServerModel.QueryAllGroup(l.ctx)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), err.Error())
	}
	nodeGroups := make([]types.ServerGroup, 0)
	tool.DeepCopy(&nodeGroups, nodeGroupList)
	return &types.GetNodeGroupListResponse{
		Total: int64(len(nodeGroups)),
		List:  nodeGroups,
	}, nil
}
