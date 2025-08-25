package server_bak

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type BatchDeleteNodeGroupLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchDeleteNodeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchDeleteNodeGroupLogic {
	return &BatchDeleteNodeGroupLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchDeleteNodeGroupLogic) BatchDeleteNodeGroup(req *types.BatchDeleteNodeGroupRequest) error {
	// Check if the group is empty
	count, err := l.svcCtx.ServerModel.QueryServerCountByServerGroups(l.ctx, req.Ids)
	if err != nil {
		l.Errorw("[BatchDeleteNodeGroup] Query Database Error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query server error: %v", err)
	}
	if count > 0 {
		return errors.Wrapf(xerr.NewErrCode(xerr.NodeGroupNotEmpty), "group is not empty")
	}
	// Delete the group
	err = l.svcCtx.ServerModel.BatchDeleteNodeGroup(l.ctx, req.Ids)
	if err != nil {
		l.Errorw("[BatchDeleteNodeGroup] Delete Database Error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), err.Error())
	}
	return nil
}
