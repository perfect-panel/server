package server

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type GetServerProtocolsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get Server Protocols
func NewGetServerProtocolsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerProtocolsLogic {
	return &GetServerProtocolsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerProtocolsLogic) GetServerProtocols(req *types.GetServerProtocolsRequest) (resp *types.GetServerProtocolsResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
