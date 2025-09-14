package server

import (
	"context"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type HasMigrateSeverNodeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewHasMigrateSeverNodeLogic Check if there is any server or node to migrate
func NewHasMigrateSeverNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HasMigrateSeverNodeLogic {
	return &HasMigrateSeverNodeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HasMigrateSeverNodeLogic) HasMigrateSeverNode() (resp *types.HasMigrateSeverNodeResponse, err error) {
	var oldCount, newCount int64
	query := l.svcCtx.DB.WithContext(l.ctx)

	err = query.Model(&server.Server{}).Count(&oldCount).Error
	if err != nil {
		l.Errorw("[HasMigrateSeverNode] Query Old Server Count Error: ", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "[HasMigrateSeverNode] Query Old Server Count Error")
	}
	err = query.Model(&node.Server{}).Count(&newCount).Error
	if err != nil {
		l.Errorw("[HasMigrateSeverNode] Query New Server Count Error: ", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "[HasMigrateSeverNode] Query New Server Count Error")
	}
	var shouldMigrate bool
	if oldCount != 0 && newCount == 0 {
		shouldMigrate = true
	}

	return &types.HasMigrateSeverNodeResponse{
		HasMigrate: shouldMigrate,
	}, nil
}
