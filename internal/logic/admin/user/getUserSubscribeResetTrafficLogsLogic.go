package user

import (
	"context"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetUserSubscribeResetTrafficLogsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user subcribe reset traffic logs
func NewGetUserSubscribeResetTrafficLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSubscribeResetTrafficLogsLogic {
	return &GetUserSubscribeResetTrafficLogsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSubscribeResetTrafficLogsLogic) GetUserSubscribeResetTrafficLogs(req *types.GetUserSubscribeResetTrafficLogsRequest) (resp *types.GetUserSubscribeResetTrafficLogsResponse, err error) {
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
			Timestamp:       content.Timestamp,
			UserSubscribeId: item.ObjectID,
		})
	}

	return &types.GetUserSubscribeResetTrafficLogsResponse{
		Total: total,
		List:  list,
	}, nil
}
