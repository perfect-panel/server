package user

import (
	"context"

	"github.com/perfect-panel/ppanel-server/pkg/constant"

	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
)

type UpdateUserPasswordLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update User Password
func NewUpdateUserPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserPasswordLogic {
	return &UpdateUserPasswordLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserPasswordLogic) UpdateUserPassword(req *types.UpdateUserPasswordRequest) error {
	userInfo := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	//update the password
	userInfo.Password = tool.EncodePassWord(req.Password)
	if err := l.svcCtx.UserModel.Update(l.ctx, userInfo); err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "Update user password error")
	}
	return nil
}
