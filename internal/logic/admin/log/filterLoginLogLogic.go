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

type FilterLoginLogLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFilterLoginLogLogic Filter login log
func NewFilterLoginLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilterLoginLogLogic {
	return &FilterLoginLogLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FilterLoginLogLogic) FilterLoginLog(req *types.FilterLoginLogRequest) (resp *types.FilterLoginLogResponse, err error) {
	data, total, err := l.svcCtx.LogModel.FilterSystemLog(l.ctx, &log.FilterParams{
		Page:     req.Page,
		Size:     req.Size,
		Type:     log.TypeLogin.Uint8(),
		ObjectID: req.UserId,
		Data:     req.Date,
		Search:   req.Search,
	})

	if err != nil {
		l.Errorf("[FilterLoginLog] failed to filter system log: %v", err.Error())
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "failed to filter system log: %v", err.Error())
	}
	var list []types.LoginLog
	for _, datum := range data {
		var item log.Login
		err = item.Unmarshal([]byte(datum.Content))
		if err != nil {
			l.Errorf("[FilterLoginLog] failed to unmarshal content: %v", err.Error())
			continue
		}
		list = append(list, types.LoginLog{
			UserId:    datum.ObjectID,
			Method:    item.Method,
			LoginIP:   item.LoginIP,
			UserAgent: item.UserAgent,
			Success:   item.Success,
			Timestamp: datum.CreatedAt.UnixMilli(),
		})
	}

	return &types.FilterLoginLogResponse{
		Total: total,
		List:  list,
	}, nil
}
