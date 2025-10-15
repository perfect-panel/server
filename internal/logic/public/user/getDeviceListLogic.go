package user

import (
	"context"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
)

type GetDeviceListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get Device List
func NewGetDeviceListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeviceListLogic {
	return &GetDeviceListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeviceListLogic) GetDeviceList() (resp *types.GetDeviceListResponse, err error) {
	userInfo := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	list, count, err := l.svcCtx.UserModel.QueryDeviceList(l.ctx, userInfo.Id)
	userRespList := make([]types.UserDevice, 0)
	tool.DeepCopy(&userRespList, list)
	resp = &types.GetDeviceListResponse{
		Total: count,
		List:  userRespList,
	}
	return
}
