package server

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type DeleteRuleGroupLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete rule group
func NewDeleteRuleGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRuleGroupLogic {
	return &DeleteRuleGroupLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRuleGroupLogic) DeleteRuleGroup(req *types.DeleteRuleGroupRequest) error {
	err := l.svcCtx.ServerModel.DeleteRuleGroup(l.ctx, req.Id)
	if err != nil {
		l.Errorw("[DeleteRuleGroup] Delete Database Error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete server rule group error: %v", err)
	}
	return nil
}
