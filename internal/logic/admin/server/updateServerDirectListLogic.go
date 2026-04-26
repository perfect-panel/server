package server

import (
	"context"
	"encoding/json"
	"fmt"

	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdateServerDirectListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateServerDirectListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateServerDirectListLogic {
	return &UpdateServerDirectListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateServerDirectListLogic) UpdateServerDirectList(req *types.UpdateServerDirectListRequest) (*types.UpdateServerDirectListResponse, error) {
	srv, err := l.svcCtx.NodeModel.FindOneServer(l.ctx, req.ServerId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find server: %v", err)
	}
	// Defensive: normalize nil array to [].
	hosts := req.DirectList
	if hosts == nil {
		hosts = []string{}
	}
	b, _ := json.Marshal(hosts)
	srv.DirectList = string(b)
	if err := l.svcCtx.NodeModel.UpdateServer(l.ctx, srv); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update: %v", err)
	}
	adminId, _ := svc.AdminIdFromCtx(l.ctx)
	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		Actor: auditmodel.ActorAdmin, ActorId: adminId,
		Action: "update_direct_list",
		Target: fmt.Sprintf("server:%d", srv.Id),
	}, map[string]interface{}{"hosts_count": len(hosts)})
	return &types.UpdateServerDirectListResponse{ServerId: srv.Id}, nil
}
