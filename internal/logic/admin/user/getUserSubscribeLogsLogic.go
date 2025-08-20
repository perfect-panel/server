package user

import (
	"context"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetUserSubscribeLogsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user subcribe logs
func NewGetUserSubscribeLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSubscribeLogsLogic {
	return &GetUserSubscribeLogsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSubscribeLogsLogic) GetUserSubscribeLogs(req *types.GetUserSubscribeLogsRequest) (resp *types.GetUserSubscribeLogsResponse, err error) {
	data, total, err := l.svcCtx.LogModel.FilterSystemLog(l.ctx, &log.FilterParams{})

	if err != nil {
		l.Errorw("[GetUserSubscribeLogs] Get User Subscribe Logs Error:", logger.Field("err", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Get User Subscribe Logs Error")
	}
	var list []types.UserSubscribeLog
	tool.DeepCopy(&list, data)

	return &types.GetUserSubscribeLogsResponse{
		List:  list,
		Total: total,
	}, err
}
