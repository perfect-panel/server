package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type EvictAliveIPLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewEvictAliveIPLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *EvictAliveIPLogic {
	return &EvictAliveIPLogic{
		Logger: logger.WithContext(ctx.Request.Context()),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// EvictAliveIP drops a single (uid, ip) from the alive ZSet on demand.
// Called by the node right after LRU-evicting that IP locally so the server
// view catches up without waiting for the score-window TTL (~3 min). If
// another node is still actively pushing this IP, its next push will
// restore it with a fresh score — that's the correct behavior.
func (l *EvictAliveIPLogic) EvictAliveIP(req *types.EvictRequest) error {
	if req.UID <= 0 || req.IP == "" {
		return nil
	}
	if err := l.svcCtx.NodeModel.RemoveAliveIP(l.ctx, req.UID, req.IP); err != nil {
		l.Errorw("[EvictAliveIP] redis remove failed",
			logger.Field("uid", req.UID),
			logger.Field("ip", req.IP),
			logger.Field("error", err.Error()))
		return err
	}
	return nil
}
