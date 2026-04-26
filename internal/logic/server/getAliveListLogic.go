package server

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/metrics"
)

type GetAliveListLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

// NewGetAliveListLogic returns the aggregated alive IP count per uid across all nodes.
// Node limiter feeds this to enforce DeviceLimit. A 2s Redis-local cache in the model
// absorbs bursty polls from many nodes.
func NewGetAliveListLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *GetAliveListLogic {
	return &GetAliveListLogic{
		Logger: logger.WithContext(ctx.Request.Context()),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAliveListLogic) GetAliveList(_ *types.GetAliveListRequest) (*types.GetAliveListResponse, error) {
	// Fetch the full IP set first. The count map is derived from it so the two
	// fields always agree. alive_detail is the source of truth; alive is kept
	// for backward compatibility with pre-v1.0.10 nodes.
	detail, err := l.svcCtx.NodeModel.AliveIPsByUID(l.ctx)
	if err != nil {
		metrics.AlivelistFetchErrorTotal.Inc()
		l.Errorw("[GetAliveList] aggregation failed", logger.Field("error", err.Error()))
		return nil, err
	}
	if detail == nil {
		detail = map[int64][]string{}
	}
	alive := make(map[int64]int64, len(detail))
	for uid, ips := range detail {
		alive[uid] = int64(len(ips))
	}
	return &types.GetAliveListResponse{Alive: alive, AliveDetail: detail}, nil
}
