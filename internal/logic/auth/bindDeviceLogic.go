package auth

import (
	"context"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BindDeviceLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBindDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindDeviceLogic {
	return &BindDeviceLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// BindDeviceToUser binds a device to a user
// If the device is already bound to another user, it will disable that user and bind the device to the current user
func (l *BindDeviceLogic) BindDeviceToUser(identifier, ip, userAgent string, currentUserId int64) error {
	if identifier == "" {
		// No device identifier provided, skip binding
		return nil
	}

	l.Infow("binding device to user",
		logger.Field("identifier", identifier),
		logger.Field("user_id", currentUserId),
		logger.Field("ip", ip),
	)

	// Check if device exists
	deviceInfo, err := l.svcCtx.UserModel.FindOneDeviceByIdentifier(l.ctx, identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Device not found, create new device record
			return l.createDeviceForUser(identifier, ip, userAgent, currentUserId)
		}
		l.Errorw("failed to query device",
			logger.Field("identifier", identifier),
			logger.Field("error", err.Error()),
		)
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query device failed: %v", err.Error())
	}

	// Device exists, check if it's bound to current user
	if deviceInfo.UserId == currentUserId {
		// Already bound to current user, just update IP and UserAgent
		l.Infow("device already bound to current user, updating info",
			logger.Field("identifier", identifier),
			logger.Field("user_id", currentUserId),
		)
		deviceInfo.Ip = ip
		deviceInfo.UserAgent = userAgent
		if err := l.svcCtx.UserModel.UpdateDevice(l.ctx, deviceInfo); err != nil {
			l.Errorw("failed to update device",
				logger.Field("identifier", identifier),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update device failed: %v", err.Error())
		}
		return nil
	}

	// Device is bound to another user, need to disable old user and rebind
	l.Infow("device bound to another user, rebinding",
		logger.Field("identifier", identifier),
		logger.Field("old_user_id", deviceInfo.UserId),
		logger.Field("new_user_id", currentUserId),
	)

	return l.rebindDeviceToNewUser(deviceInfo, ip, userAgent, currentUserId)
}

func (l *BindDeviceLogic) createDeviceForUser(identifier, ip, userAgent string, userId int64) error {
	l.Infow("creating new device for user",
		logger.Field("identifier", identifier),
		logger.Field("user_id", userId),
	)

	err := l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		// Create device auth method
		authMethod := &user.AuthMethods{
			UserId:         userId,
			AuthType:       "device",
			AuthIdentifier: identifier,
			Verified:       true,
		}
		if err := db.Create(authMethod).Error; err != nil {
			l.Errorw("failed to create device auth method",
				logger.Field("user_id", userId),
				logger.Field("identifier", identifier),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create device auth method failed: %v", err)
		}

		// Create device record
		deviceInfo := &user.Device{
			Ip:         ip,
			UserId:     userId,
			UserAgent:  userAgent,
			Identifier: identifier,
			Enabled:    true,
			Online:     false,
		}
		if err := db.Create(deviceInfo).Error; err != nil {
			l.Errorw("failed to create device",
				logger.Field("user_id", userId),
				logger.Field("identifier", identifier),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create device failed: %v", err)
		}

		return nil
	})

	if err != nil {
		l.Errorw("device creation failed",
			logger.Field("identifier", identifier),
			logger.Field("user_id", userId),
			logger.Field("error", err.Error()),
		)
		return err
	}

	l.Infow("device created successfully",
		logger.Field("identifier", identifier),
		logger.Field("user_id", userId),
	)

	return nil
}

func (l *BindDeviceLogic) rebindDeviceToNewUser(deviceInfo *user.Device, ip, userAgent string, newUserId int64) error {
	oldUserId := deviceInfo.UserId

	err := l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		// Check if old user has other auth methods besides device
		var authMethods []user.AuthMethods
		if err := db.Where("user_id = ?", oldUserId).Find(&authMethods).Error; err != nil {
			l.Errorw("failed to query auth methods for old user",
				logger.Field("old_user_id", oldUserId),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query auth methods failed: %v", err)
		}

		// Count non-device auth methods
		nonDeviceAuthCount := 0
		for _, auth := range authMethods {
			if auth.AuthType != "device" {
				nonDeviceAuthCount++
			}
		}

		// Only disable old user if they have no other auth methods
		if nonDeviceAuthCount == 0 {
			falseVal := false
			if err := db.Model(&user.User{}).Where("id = ?", oldUserId).Update("enable", &falseVal).Error; err != nil {
				l.Errorw("failed to disable old user",
					logger.Field("old_user_id", oldUserId),
					logger.Field("error", err.Error()),
				)
				return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "disable old user failed: %v", err)
			}

			l.Infow("disabled old user (no other auth methods)",
				logger.Field("old_user_id", oldUserId),
			)
		} else {
			l.Infow("old user has other auth methods, not disabling",
				logger.Field("old_user_id", oldUserId),
				logger.Field("non_device_auth_count", nonDeviceAuthCount),
			)
		}

		// Update device auth method to new user
		if err := db.Model(&user.AuthMethods{}).
			Where("auth_type = ? AND auth_identifier = ?", "device", deviceInfo.Identifier).
			Update("user_id", newUserId).Error; err != nil {
			l.Errorw("failed to update device auth method",
				logger.Field("identifier", deviceInfo.Identifier),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update device auth method failed: %v", err)
		}

		// Update device record
		deviceInfo.UserId = newUserId
		deviceInfo.Ip = ip
		deviceInfo.UserAgent = userAgent
		deviceInfo.Enabled = true

		if err := db.Save(deviceInfo).Error; err != nil {
			l.Errorw("failed to update device",
				logger.Field("identifier", deviceInfo.Identifier),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update device failed: %v", err)
		}

		return nil
	})

	if err != nil {
		l.Errorw("device rebinding failed",
			logger.Field("identifier", deviceInfo.Identifier),
			logger.Field("old_user_id", oldUserId),
			logger.Field("new_user_id", newUserId),
			logger.Field("error", err.Error()),
		)
		return err
	}

	l.Infow("device rebound successfully",
		logger.Field("identifier", deviceInfo.Identifier),
		logger.Field("old_user_id", oldUserId),
		logger.Field("new_user_id", newUserId),
	)

	return nil
}
