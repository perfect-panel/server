package server

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
)

type QueryServerProtocolConfigLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewQueryServerProtocolConfigLogic Get Server Protocol Config
func NewQueryServerProtocolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryServerProtocolConfigLogic {
	return &QueryServerProtocolConfigLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryServerProtocolConfigLogic) QueryServerProtocolConfig(req *types.QueryServerConfigRequest) (resp *types.QueryServerConfigResponse, err error) {
	// find server
	data, err := l.svcCtx.NodeModel.FindOneServer(l.ctx, req.ServerID)
	if err != nil {
		l.Errorf("[GetServerProtocols] FindOneServer Error: %s", err.Error())
		return nil, err
	}

	// handler protocols
	var protocols []types.Protocol
	dst, err := data.UnmarshalProtocols()
	if err != nil {
		l.Errorf("[FilterServerList] UnmarshalProtocols Error: %s", err.Error())
		return nil, err
	}
	tool.DeepCopy(&protocols, dst)

	return &types.QueryServerConfigResponse{
		Protocols: protocols,
	}, nil
}
