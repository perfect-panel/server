package server

import (
	"context"

	"github.com/perfect-panel/ppanel-server/internal/model/server"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"

	"github.com/pkg/errors"
)

type CreateNodeGroupLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateNodeGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNodeGroupLogic {
	return &CreateNodeGroupLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateNodeGroupLogic) CreateNodeGroup(req *types.CreateNodeGroupRequest) error {
	groupInfo := &server.Group{
		Name:        req.Name,
		Description: req.Description,
	}
	err := l.svcCtx.ServerModel.InsertGroup(l.ctx, groupInfo)
	if err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), err.Error())
	}
	return nil
}
