package user

import (
	"context"
	"fmt"

	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type DeleteCurrentUserAccountLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete Current User Account
func NewDeleteCurrentUserAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCurrentUserAccountLogic {
	return &DeleteCurrentUserAccountLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCurrentUserAccountLogic) DeleteCurrentUserAccount() (err error) {
	userInfo, exists := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !exists {
		return nil
	}

	userInfo, err = l.svcCtx.UserModel.FindOne(l.ctx, userInfo.Id)
	if err != nil {
		l.Errorw("FindOne Error", logger.Field("error", err))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find user auth methods failed: %v", err.Error())
	}

	err = l.svcCtx.UserModel.Transaction(l.ctx, func(tx *gorm.DB) error {
		//delete user devices
		if len(userInfo.UserDevices) > 0 {
			for _, device := range userInfo.UserDevices {
				if err = l.svcCtx.UserModel.DeleteDevice(l.ctx, device.Id, tx); err != nil {
					return err
				}
			}
		}

		// delete user auth methods
		if len(userInfo.AuthMethods) > 0 {
			for _, authMethod := range userInfo.AuthMethods {
				if err = l.svcCtx.UserModel.DeleteUserAuthMethods(l.ctx, userInfo.Id, authMethod.AuthType); err != nil {
					return err
				}
			}
		}

		// delete user subscribes
		var subscribes []*user.SubscribeDetails
		subscribes, err = l.svcCtx.UserModel.QueryUserSubscribe(l.ctx, userInfo.Id)
		if err != nil {
			return err
		}
		for _, subscribe := range subscribes {
			if err = l.svcCtx.UserModel.DeleteSubscribe(l.ctx, subscribe.Token, tx); err != nil {
				return err
			}
		}
		// delete user account
		return l.svcCtx.UserModel.BatchDeleteUser(l.ctx, []int64{userInfo.Id}, tx)
	})
	if err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "find user auth methods failed: %v", err.Error())
	}
	sessionIdCacheKey := fmt.Sprintf("%v:%v", config.SessionIdKey, l.ctx.Value(constant.CtxKeySessionID))
	if err = l.svcCtx.Redis.Del(l.ctx, sessionIdCacheKey).Err(); err != nil {
		l.Logger.Errorf("delete session id cache failed: %v", err.Error())
	}
	return

}
