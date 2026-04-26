package user

import (
	"context"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/metrics"
)

// MaxBatchOnlineStatusUIDs caps one admin list-page query. Pages larger than
// this should paginate rather than single-shot.
const MaxBatchOnlineStatusUIDs = 500

type BatchGetUserOnlineStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logger.Logger
}

func NewBatchGetUserOnlineStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUserOnlineStatusLogic {
	return &BatchGetUserOnlineStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logger.WithContext(ctx),
	}
}

// BatchGetUserOnlineStatus returns online IP counts + 24h reject counts for
// many uids in one pass (HGETALL + single alivelist read). Used by the admin
// user-list page to render two summary columns without N+1 lookups.
func (l *BatchGetUserOnlineStatusLogic) BatchGetUserOnlineStatus(req *types.BatchGetUserOnlineStatusRequest) (*types.BatchGetUserOnlineStatusResponse, error) {
	metrics.AdminOnlineStatusQPS.Inc()
	resp := &types.BatchGetUserOnlineStatusResponse{Items: []types.UserOnlineStatus{}}
	if len(req.UIDs) == 0 {
		return resp, nil
	}
	uids := req.UIDs
	if len(uids) > MaxBatchOnlineStatusUIDs {
		l.Errorw("[BatchGetUserOnlineStatus] payload truncated",
			logger.Field("received", len(uids)),
			logger.Field("kept", MaxBatchOnlineStatusUIDs))
		uids = uids[:MaxBatchOnlineStatusUIDs]
	}

	// alivelist hits the 2s local cache, so re-using it N times is cheap enough.
	alive, err := l.svcCtx.NodeModel.AliveListByUID(l.ctx)
	if err != nil {
		l.Errorw("[BatchGetUserOnlineStatus] alivelist failed", logger.Field("error", err.Error()))
		alive = map[int64]int64{}
	}
	rejects, err := l.svcCtx.NodeModel.RejectCount24hBatch(l.ctx, uids)
	if err != nil {
		l.Errorw("[BatchGetUserOnlineStatus] reject batch failed", logger.Field("error", err.Error()))
		rejects = map[int64]int64{}
	}

	items := make([]types.UserOnlineStatus, 0, len(uids))
	for _, uid := range uids {
		items = append(items, types.UserOnlineStatus{
			UID:               uid,
			OnlineIPs:         nil, // list view does not expose raw IPs; detail endpoint does.
			OnlineIPCount:     alive[uid],
			RejectCount24h:    rejects[uid],
			DataSourceVersion: MinNodeVersionWithRejectReport,
		})
	}
	resp.Items = items
	return resp, nil
}
