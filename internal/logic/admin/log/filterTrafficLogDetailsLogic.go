package log

import (
	"context"
	"time"

	"github.com/perfect-panel/server/internal/model/traffic"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type FilterTrafficLogDetailsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFilterTrafficLogDetailsLogic Filter traffic log details
func NewFilterTrafficLogDetailsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilterTrafficLogDetailsLogic {
	return &FilterTrafficLogDetailsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FilterTrafficLogDetailsLogic) FilterTrafficLogDetails(req *types.FilterTrafficLogDetailsRequest) (resp *types.FilterTrafficLogDetailsResponse, err error) {
	var start, end time.Time
	if req.Date != "" {
		day, err := time.ParseInLocation("2006-01-02", req.Date, time.Local)
		if err != nil {
			l.Errorw("[FilterTrafficLogDetails] Date Parse Error", logger.Field("error", err.Error()))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), " date parse error: %s", err.Error())
		}
		start = day
		end = day.Add(24*time.Hour - time.Nanosecond)
	} else {
		// query today
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.Add(24*time.Hour - time.Nanosecond)
	}
	var data []*traffic.TrafficLog
	tx := l.svcCtx.DB.WithContext(l.ctx).Model(&traffic.TrafficLog{})
	if req.ServerId != 0 {
		tx = tx.Where("server_id = ?", req.ServerId)
	}
	if !start.IsZero() && !end.IsZero() {
		tx = tx.Where("timestamp BETWEEN ? AND ?", start, end)
	}
	if req.UserId != 0 {
		tx = tx.Where("user_id = ?", req.UserId)
	}
	if req.SubscribeId != 0 {
		tx = tx.Where("subscribe_id = ?", req.SubscribeId)
	}
	var total int64
	err = tx.Count(&total).Limit(req.Size).Offset((req.Page - 1) * req.Size).Find(&data).Error
	if err != nil {
		l.Errorw("[FilterTrafficLogDetails] Query Database Error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), " database query error: %s", err.Error())
	}

	var logs []types.TrafficLogDetails
	for _, v := range data {
		logs = append(logs, types.TrafficLogDetails{
			Id:          v.Id,
			UserId:      v.UserId,
			ServerId:    v.ServerId,
			SubscribeId: v.SubscribeId,
			Download:    v.Download,
			Upload:      v.Upload,
			Timestamp:   v.Timestamp.UnixMilli(),
		})
	}

	return &types.FilterTrafficLogDetailsResponse{
		List:  logs,
		Total: total,
	}, nil
}
