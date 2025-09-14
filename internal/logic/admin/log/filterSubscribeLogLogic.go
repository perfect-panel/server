package log

import (
	"context"
	"strconv"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type FilterSubscribeLogLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFilterSubscribeLogLogic Filter subscribe log
func NewFilterSubscribeLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilterSubscribeLogLogic {
	return &FilterSubscribeLogLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FilterSubscribeLogLogic) FilterSubscribeLog(req *types.FilterSubscribeLogRequest) (resp *types.FilterSubscribeLogResponse, err error) {
	params := &log.FilterParams{
		Page:     req.Page,
		Size:     req.Size,
		Type:     log.TypeSubscribe.Uint8(),
		Data:     req.Date,
		ObjectID: req.UserId,
	}

	if req.UserSubscribeId != 0 {
		params.Search = `"user_subscribe_id":` + strconv.FormatInt(req.UserSubscribeId, 10)
	}

	data, total, err := l.svcCtx.LogModel.FilterSystemLog(l.ctx, params)
	if err != nil {
		l.Errorf("[FilterSubscribeLog] failed to filter system log: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "failed to filter system log")
	}

	var list []types.SubscribeLog
	for _, datum := range data {
		var content log.Subscribe
		err = content.Unmarshal([]byte(datum.Content))
		if err != nil {
			l.Errorf("[FilterSubscribeLog] failed to unmarshal content: %v", err.Error())
			continue
		}
		list = append(list, types.SubscribeLog{
			UserId:          datum.ObjectID,
			Token:           content.Token,
			UserAgent:       content.UserAgent,
			ClientIP:        content.ClientIP,
			UserSubscribeId: content.UserSubscribeId,
			Timestamp:       datum.CreatedAt.UnixMilli(),
		})
	}

	return &types.FilterSubscribeLogResponse{
		Total: total,
		List:  list,
	}, nil
}
