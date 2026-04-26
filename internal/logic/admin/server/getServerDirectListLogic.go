package server

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetServerDirectListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetServerDirectListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerDirectListLogic {
	return &GetServerDirectListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerDirectListLogic) GetServerDirectList(req *types.GetServerDirectListRequest) (*types.GetServerDirectListResponse, error) {
	srv, err := l.svcCtx.NodeModel.FindOneServer(l.ctx, req.ServerId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find server: %v", err)
	}
	out := &types.GetServerDirectListResponse{ServerId: srv.Id}
	if srv.DirectList != "" {
		_ = json.Unmarshal([]byte(srv.DirectList), &out.DirectList)
	}
	return out, nil
}
