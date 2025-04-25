package user

import (
	"context"

	"github.com/perfect-panel/ppanel-server/pkg/constant"

	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdatePasswordLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update Password
func NewUpdatePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePasswordLogic {
	return &UpdatePasswordLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePasswordLogic) UpdatePassword(req *types.UpdatePasswordRequeset) error {
	userInfo := l.ctx.Value(constant.CtxKeyUser).(*user.User)

	// Verify password
	if !tool.VerifyPassWord(req.Password, userInfo.Password) {
		return errors.Wrapf(xerr.NewErrCode(xerr.UserPasswordError), "user password")
	}
	userInfo.Password = tool.EncodePassWord(req.NewPassword)
	err := l.svcCtx.UserModel.Update(l.ctx, userInfo)
	if err != nil {
		l.Errorw("update user password error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update user password")
	}
	return err
}
