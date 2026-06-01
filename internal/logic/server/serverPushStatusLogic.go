package server

import (
	"context"
	"errors"
	"time"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type ServerPushStatusLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewServerPushStatusLogic Push server status
func NewServerPushStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ServerPushStatusLogic {
	return &ServerPushStatusLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ServerPushStatusLogic) ServerPushStatus(req *types.ServerPushStatusRequest) error {
	// Find server info
	nodeStore := l.svcCtx.Store.Node()
	serverInfo, err := nodeStore.FindOneServer(l.ctx, req.ServerId)
	if err != nil || serverInfo.Id <= 0 {
		l.Errorw("[PushOnlineUsers] FindOne error", logger.Field("error", err))
		return errors.New("server not found")
	}
	err = nodeStore.UpdateStatusCache(l.ctx, req.ServerId, &node.Status{
		Cpu:       req.Cpu,
		Mem:       req.Mem,
		Disk:      req.Disk,
		UpdatedAt: req.UpdatedAt,
	})
	if err != nil {
		l.Errorw("[ServerPushStatus] UpdateNodeStatus error", logger.Field("error", err))
		return errors.New("update node status failed")
	}
	now := time.Now()
	serverInfo.LastReportedAt = &now
	certFingerprintChanged := false
	if req.CertFingerprintSha256 != "" {
		protocols, err := serverInfo.UnmarshalProtocols()
		if err != nil {
			l.Errorw("[ServerPushStatus] UnmarshalProtocols error", logger.Field("error", err.Error()))
			return errors.New("unmarshal server protocols failed")
		}
		if updateReportedCertFingerprintSha256(protocols, req.Protocol, req.CertFingerprintSha256) {
			if err = serverInfo.MarshalProtocols(protocols); err != nil {
				l.Errorw("[ServerPushStatus] MarshalProtocols error", logger.Field("error", err.Error()))
				return errors.New("marshal server protocols failed")
			}
			certFingerprintChanged = true
		}
	}

	err = nodeStore.UpdateServer(l.ctx, serverInfo)
	if err != nil {
		l.Errorw("[ServerPushStatus] UpdateServer error", logger.Field("error", err))
		return nil
	}
	if certFingerprintChanged {
		if err = nodeStore.ClearNodeCache(l.ctx, &node.FilterNodeParams{
			Page:     1,
			Size:     1000,
			ServerId: []int64{req.ServerId},
		}); err != nil {
			l.Errorw("[ServerPushStatus] ClearNodeCache error", logger.Field("error", err.Error()))
		}
	}

	return nil
}
