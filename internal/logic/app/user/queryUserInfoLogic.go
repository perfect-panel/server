package user

import (
	"context"

	"github.com/perfect-panel/server/pkg/constant"

	"github.com/perfect-panel/server/internal/model/user"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type QueryUserInfoLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// query user info
func NewQueryUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserInfoLogic {
	return &QueryUserInfoLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryUserInfoLogic) QueryUserInfo() (resp *types.UserInfoResponse, err error) {
	u := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	var devices []types.UserDevice
	if len(u.UserDevices) != 0 {
		for _, device := range u.UserDevices {
			devices = append(devices, types.UserDevice{
				Id:         device.Id,
				Identifier: device.Identifier,
				Online:     device.Online,
			})
		}
	}
	var authMeths []types.UserAuthMethod
	authMethods, err := l.svcCtx.UserModel.FindUserAuthMethods(l.ctx, u.Id)
	if err == nil && len(authMeths) != 0 {
		for _, as := range authMethods {
			authMeths = append(authMeths, types.UserAuthMethod{
				AuthType:       as.AuthType,
				AuthIdentifier: as.AuthIdentifier,
			})
		}
	}

	resp = &types.UserInfoResponse{
		Id:          u.Id,
		Balance:     u.Balance,
		Avatar:      u.Avatar,
		ReferCode:   u.ReferCode,
		RefererId:   u.RefererId,
		Devices:     devices,
		AuthMethods: authMeths,
	}
	return
}
