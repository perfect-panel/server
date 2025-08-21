package user

import (
	"context"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type ResetSubscribeTrafficLogLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewResetSubscribeTrafficLogLogic Reset Subscribe Traffic Log
func NewResetSubscribeTrafficLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetSubscribeTrafficLogLogic {
	return &ResetSubscribeTrafficLogLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetSubscribeTrafficLogLogic) ResetSubscribeTrafficLog(req *types.ResetSubscribeTrafficLogRequest) (resp *types.ResetSubscribeTrafficLogResponse, err error) {
	data, total, err := l.svcCtx.LogModel.FilterSystemLog(l.ctx, &log.FilterParams{
		Page:     req.Page,
		Size:     req.Size,
		Type:     log.TypeResetSubscribe.Uint8(),
		ObjectID: req.UserSubscribeId,
	})
	if err != nil {
		l.Errorf("[ResetSubscribeTrafficLog] failed to filter system log: %v", err)
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FilterSystemLog failed, err: %v", err)
	}

	var list []types.ResetSubscribeTrafficLog

	for _, item := range data {
		var content log.ResetSubscribe
		if err = content.Unmarshal([]byte(item.Content)); err != nil {
			l.Errorf("[ResetSubscribeTrafficLog] failed to unmarshal log: %v", err)
			continue
		}
		list = append(list, types.ResetSubscribeTrafficLog{
			Id:              item.Id,
			Type:            content.Type,
			OrderNo:         content.OrderNo,
			ResetAt:         content.ResetAt,
			UserSubscribeId: item.ObjectID,
		})
	}

	return &types.ResetSubscribeTrafficLogResponse{
		Total: total,
		List:  list,
	}, nil
}
