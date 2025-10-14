package user

import (
	"context"
	"fmt"

	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type UnbindDeviceLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Unbind Device
func NewUnbindDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbindDeviceLogic {
	return &UnbindDeviceLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnbindDeviceLogic) UnbindDevice(req *types.UnbindDeviceRequest) error {
	userInfo := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	device, err := l.svcCtx.UserModel.FindOneDevice(l.ctx, req.Id)
	if err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DeviceNotExist), "find device")
	}

	if device.UserId != userInfo.Id {
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "device not belong to user")
	}

	return l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var deleteDevice user.Device
		err = tx.Model(&deleteDevice).Where("id = ?", req.Id).First(&deleteDevice).Error
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.QueueEnqueueError), "find device err: %v", err)
		}
		err = tx.Delete(deleteDevice).Error
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete device err: %v", err)
		}
		var userAuth user.AuthMethods
		err = tx.Model(&userAuth).Where("auth_identifier = ? and auth_type = ?", deleteDevice.Identifier, "device").First(&userAuth).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find device online record err: %v", err)
		}

		err = tx.Delete(&userAuth).Error
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete device online record err: %v", err)
		}
		sessionId := l.ctx.Value(constant.CtxKeySessionID)
		sessionIdCacheKey := fmt.Sprintf("%v:%v", config.SessionIdKey, sessionId)
		l.svcCtx.Redis.Del(l.ctx, sessionIdCacheKey)
		return nil
	})
}
