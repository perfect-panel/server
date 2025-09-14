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

type FilterResetSubscribeLogLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFilterResetSubscribeLogLogic Filter reset subscribe log
func NewFilterResetSubscribeLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilterResetSubscribeLogLogic {
	return &FilterResetSubscribeLogLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FilterResetSubscribeLogLogic) FilterResetSubscribeLog(req *types.FilterResetSubscribeLogRequest) (resp *types.FilterResetSubscribeLogResponse, err error) {
	data, total, err := l.svcCtx.LogModel.FilterSystemLog(l.ctx, &log.FilterParams{
		Page:     req.Page,
		Size:     req.Size,
		Type:     log.TypeResetSubscribe.Uint8(),
		ObjectID: req.UserSubscribeId,
		Data:     req.Date,
		Search:   req.Search,
	})

	if err != nil {
		l.Errorf("[FilterResetSubscribeLog] failed to filter system log: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "failed to filter system log: %v", err.Error())
	}

	var list []types.ResetSubscribeLog

	for _, item := range data {
		var content log.ResetSubscribe
		err = content.Unmarshal([]byte(item.Content))
		if err != nil {
			l.Errorf("[FilterResetSubscribeLog] failed to unmarshal content: %v", err.Error())
			continue
		}
		list = append(list, types.ResetSubscribeLog{
			Type:            content.Type,
			UserId:          content.UserId,
			UserSubscribeId: item.ObjectID,
			OrderNo:         content.OrderNo,
			Timestamp:       content.Timestamp,
		})
	}

	return &types.FilterResetSubscribeLogResponse{
		List:  list,
		Total: total,
	}, nil
}
