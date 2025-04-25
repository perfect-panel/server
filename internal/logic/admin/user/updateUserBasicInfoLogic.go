package user

import (
	"context"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdateUserBasicInfoLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateUserBasicInfoLogic Update user basic info
func NewUpdateUserBasicInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBasicInfoLogic {
	return &UpdateUserBasicInfoLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserBasicInfoLogic) UpdateUserBasicInfo(req *types.UpdateUserBasiceInfoRequest) error {
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, req.UserId)
	if err != nil {
		l.Errorw("[UpdateUserBasicInfoLogic] Find User Error:", logger.Field("err", err.Error()), logger.Field("userId", req.UserId))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Find User Error")
	}

	tool.DeepCopy(userInfo, req)
	if req.Avatar != "" && !tool.IsValidImageSize(req.Avatar, 1024) {
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Invalid Image Size")
	}
	if req.Password != "" {
		l.Infow("[UpdateUserBasicInfoLogic] Update User Password:", logger.Field("userId", req.UserId), logger.Field("password", req.Password))
		userInfo.Password = tool.EncodePassWord(req.Password)
	}

	err = l.svcCtx.UserModel.Update(l.ctx, userInfo)
	if err != nil {
		l.Errorw("[UpdateUserBasicInfoLogic] Update User Error:", logger.Field("err", err.Error()), logger.Field("userId", req.UserId))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "Update User Error")
	}

	return nil
}
