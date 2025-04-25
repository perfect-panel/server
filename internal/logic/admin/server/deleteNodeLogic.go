package server

import (
	"context"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type DeleteNodeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteNodeLogic {
	return &DeleteNodeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteNodeLogic) DeleteNode(req *types.DeleteNodeRequest) error {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// Delete server
		err := l.svcCtx.ServerModel.Delete(l.ctx, req.Id)
		if err != nil {
			return err
		}
		// Delete server to subscribe
		subs, err := l.svcCtx.SubscribeModel.QuerySubscribeIdsByServerIdAndServerGroupId(l.ctx, req.Id, 0)
		if err != nil {
			return err
		}
		for _, sub := range subs {
			servers := tool.StringToInt64Slice(sub.Server)
			newServers := tool.RemoveElementBySlice(servers, req.Id)
			sub.Server = tool.Int64SliceToString(newServers)
			if err = l.svcCtx.SubscribeModel.Update(l.ctx, sub); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		l.Errorw("[DeleteNode] Delete Database Error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete server error: %v", err)
	}
	return nil
}
