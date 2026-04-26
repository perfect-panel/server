package user

import (
	"context"

	userModel "github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/metrics"
)

// MinNodeVersionWithRejectReport is advertised back to clients so the UI can
// show "statistics cover node v1.0.9+ only" during rolling upgrades.
const MinNodeVersionWithRejectReport = "v1.0.9+"

type GetUserOnlineStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger.Logger
}

func NewGetUserOnlineStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOnlineStatusLogic {
	return &GetUserOnlineStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logger.WithContext(ctx),
	}
}

// GetUserOnlineStatus returns limiter-visible state for a single uid.
// Admin access is logged as a lightweight audit trail until a dedicated
// audit-log table lands; fine-grained RBAC permission `user:view_online_ip`
// is deferred (currently relies on existing IsAdmin gate in AuthMiddleware).
func (l *GetUserOnlineStatusLogic) GetUserOnlineStatus(req *types.GetUserOnlineStatusRequest) (*types.UserOnlineStatus, error) {
	metrics.AdminOnlineStatusQPS.Inc()
	resp := &types.UserOnlineStatus{
		UID:               req.UID,
		OnlineIPs:         []string{},
		DataSourceVersion: MinNodeVersionWithRejectReport,
	}
	if req.UID <= 0 {
		return resp, nil
	}

	ips, err := l.svcCtx.NodeModel.ListOnlineIPsByUID(l.ctx, req.UID)
	if err != nil {
		l.Errorw("[GetUserOnlineStatus] list online IPs failed",
			logger.Field("uid", req.UID),
			logger.Field("error", err.Error()))
		return resp, err
	}
	resp.OnlineIPs = ips
	resp.OnlineIPCount = int64(len(ips))

	rejectCount, err := l.svcCtx.NodeModel.RejectCount24hByUID(l.ctx, req.UID)
	if err != nil {
		l.Errorw("[GetUserOnlineStatus] reject count failed",
			logger.Field("uid", req.UID),
			logger.Field("error", err.Error()))
	}
	resp.RejectCount24h = rejectCount

	// Audit trail via log until a persistent audit store is introduced.
	adminID := adminIDFromCtx(l.ctx)
	l.Infow("[audit] admin viewed online status",
		logger.Field("admin_id", adminID),
		logger.Field("target_uid", req.UID))

	return resp, nil
}

// adminIDFromCtx extracts the current admin's user id from the context injected
// by AuthMiddleware; returns 0 if absent (e.g. misconfigured route).
func adminIDFromCtx(ctx context.Context) int64 {
	if u, ok := ctx.Value(constant.CtxKeyUser).(*userModel.User); ok && u != nil {
		return u.Id
	}
	return 0
}
