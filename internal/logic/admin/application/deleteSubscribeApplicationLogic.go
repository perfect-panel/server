package application

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type DeleteSubscribeApplicationLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDeleteSubscribeApplicationLogic Delete subscribe application
func NewDeleteSubscribeApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSubscribeApplicationLogic {
	return &DeleteSubscribeApplicationLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSubscribeApplicationLogic) DeleteSubscribeApplication(req *types.DeleteSubscribeApplicationRequest) error {
	err := l.svcCtx.ClientModel.Delete(l.ctx, req.Id)
	if err != nil {
		l.Errorf("Failed to delete subscribe application with ID %d: %v", req.Id, err)
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), err.Error())
	}
	return nil
}
