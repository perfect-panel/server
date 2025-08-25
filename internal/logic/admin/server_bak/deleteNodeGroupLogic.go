package server_bak

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type DeleteNodeGroupLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteNodeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteNodeGroupLogic {
	return &DeleteNodeGroupLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteNodeGroupLogic) DeleteNodeGroup(req *types.DeleteNodeGroupRequest) error {
	// Check if the group is empty
	count, err := l.svcCtx.ServerModel.QueryServerCountByServerGroups(l.ctx, []int64{req.Id})
	if err != nil {
		l.Errorw("[DeleteNodeGroup] Query Database Error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query server error: %v", err)
	}
	if count > 0 {
		return errors.Wrapf(xerr.NewErrCode(xerr.NodeGroupNotEmpty), "group is not empty")
	}
	// Delete the group
	err = l.svcCtx.ServerModel.DeleteGroup(l.ctx, req.Id)
	if err != nil {
		l.Errorw("[DeleteNodeGroup] Delete Database Error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), err.Error())
	}
	return nil
}
