package log

import (
	"context"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type FilterCommissionLogLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFilterCommissionLogLogic Filter commission log
func NewFilterCommissionLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilterCommissionLogLogic {
	return &FilterCommissionLogLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FilterCommissionLogLogic) FilterCommissionLog(req *types.FilterCommissionLogRequest) (resp *types.FilterCommissionLogResponse, err error) {
	data, total, err := l.svcCtx.LogModel.FilterSystemLog(l.ctx, &log.FilterParams{
		Page:     req.Page,
		Size:     req.Size,
		Data:     req.Date,
		Type:     log.TypeCommission.Uint8(),
		ObjectID: req.UserId,
	})
	if err != nil {
		l.Errorw("Query User Commission Log failed", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Query User Commission Log failed")
	}
	var list []types.CommissionLog

	for _, datum := range data {
		var content log.Commission
		if err = content.Unmarshal([]byte(datum.Content)); err != nil {
			l.Errorf("unmarshal commission log content failed: %v", err.Error())
			continue
		}
		list = append(list, types.CommissionLog{
			UserId:    datum.ObjectID,
			Type:      content.Type,
			Amount:    content.Amount,
			OrderNo:   content.OrderNo,
			Timestamp: content.Timestamp,
		})
	}
	return &types.FilterCommissionLogResponse{
		Total: total,
		List:  list,
	}, nil
}
