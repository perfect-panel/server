package user

import (
	"context"

	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetUserLoginLogsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get user login logs
func NewGetUserLoginLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLoginLogsLogic {
	return &GetUserLoginLogsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLoginLogsLogic) GetUserLoginLogs(req *types.GetUserLoginLogsRequest) (resp *types.GetUserLoginLogsResponse, err error) {
	data, total, err := l.svcCtx.UserModel.FilterLoginLogList(l.ctx, req.Page, req.Size, &user.LoginLogFilterParams{
		UserId: req.UserId,
	})
	if err != nil {
		l.Errorw("[GetUserLoginLogs] get user login logs failed", logger.Field("error", err.Error()), logger.Field("request", req))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get user login logs failed: %v", err.Error())
	}
	var list []types.UserLoginLog
	tool.DeepCopy(&list, data)
	return &types.GetUserLoginLogsResponse{
		Total: total,
		List:  list,
	}, nil
}
