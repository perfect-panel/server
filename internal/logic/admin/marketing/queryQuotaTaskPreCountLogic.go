package marketing

import (
	"context"
	"time"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type QueryQuotaTaskPreCountLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewQueryQuotaTaskPreCountLogic Query quota task pre-count
func NewQueryQuotaTaskPreCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryQuotaTaskPreCountLogic {
	return &QueryQuotaTaskPreCountLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryQuotaTaskPreCountLogic) QueryQuotaTaskPreCount(req *types.QueryQuotaTaskPreCountRequest) (resp *types.QueryQuotaTaskPreCountResponse, err error) {
	tx := l.svcCtx.DB.WithContext(l.ctx).Model(&user.Subscribe{})
	var count int64

	if len(req.Subscribers) > 0 {
		tx = tx.Where("`subscribe_id` IN ?", req.Subscribers)
	}

	if req.IsActive != nil && *req.IsActive {
		tx = tx.Where("`status` IN ?", []int64{0, 1, 2}) // 0: Pending 1: Active 2: Finished
	}
	if req.StartTime != 0 {
		start := time.UnixMilli(req.StartTime)
		tx = tx.Where("`start_time` <= ?", start)
	}
	if req.EndTime != 0 {
		end := time.UnixMilli(req.EndTime)
		tx = tx.Where("`expire_time` >= ?", end)
	}
	if err = tx.Count(&count).Error; err != nil {
		l.Errorf("[QueryQuotaTaskPreCount] count error: %v", err.Error())
		return nil, err
	}

	return &types.QueryQuotaTaskPreCountResponse{
		Count: count,
	}, nil
}
