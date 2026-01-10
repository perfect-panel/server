package auth

import (
	"context"
	"time"

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

	// Check device limit
	deviceLimit := l.svcCtx.Config.Register.DeviceLimit
	if deviceLimit > 0 {
		// Count current user's devices
		var deviceCount int64
		if err := l.svcCtx.DB.Model(&user.Device{}).Where("user_id = ?", userId).Count(&deviceCount).Error; err != nil {
			l.Errorw("failed to count user devices",
				logger.Field("user_id", userId),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "count devices failed: %v", err.Error())
		}

		// Check if limit reached
		if deviceCount >= deviceLimit {
			l.Errorw("device limit reached",
				logger.Field("user_id", userId),
				logger.Field("device_count", deviceCount),
				logger.Field("device_limit", deviceLimit),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "device limit reached: maximum %d devices allowed", deviceLimit)
		}

		l.Infow("device limit check passed",
			logger.Field("user_id", userId),
			logger.Field("device_count", deviceCount),
			logger.Field("device_limit", deviceLimit),
		)
	}

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
			Ip:     ip,
			UserId: userId,

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

	// Check device limit for new user
	deviceLimit := l.svcCtx.Config.Register.DeviceLimit
	if deviceLimit > 0 {
		// Count new user's current devices (excluding the one being rebound)
		var deviceCount int64
		if err := l.svcCtx.DB.Model(&user.Device{}).Where("user_id = ? AND id != ?", newUserId, deviceInfo.Id).Count(&deviceCount).Error; err != nil {
			l.Errorw("failed to count new user devices",
				logger.Field("user_id", newUserId),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "count devices failed: %v", err.Error())
		}

		// Check if limit reached
		if deviceCount >= deviceLimit {
			l.Errorw("device limit reached for new user",
				logger.Field("user_id", newUserId),
				logger.Field("device_count", deviceCount),
				logger.Field("device_limit", deviceLimit),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "device limit reached: maximum %d devices allowed", deviceLimit)
		}

		l.Infow("device limit check passed for rebinding",
			logger.Field("user_id", newUserId),
			logger.Field("device_count", deviceCount),
			logger.Field("device_limit", deviceLimit),
		)
	}

	var users []*user.User
	err := l.svcCtx.DB.Where("id in (?)", []int64{oldUserId, newUserId}).Find(&users).Error
	if err != nil {
		l.Errorw("failed to query users for rebinding",
			logger.Field("old_user_id", oldUserId),
			logger.Field("new_user_id", newUserId),
			logger.Field("error", err.Error()),
		)
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query users failed: %v", err)
	}
	err = l.svcCtx.UserModel.Transaction(l.ctx, func(tx *gorm.DB) error {
		//检查旧设备是否存在认证方式
		var authMethod user.AuthMethods
		err := tx.Where("auth_type = ? AND auth_identifier = ?", "device", deviceInfo.Identifier).Find(&authMethod).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorw("failed to query device auth method",
				logger.Field("identifier", deviceInfo.Identifier),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query device auth method failed: %v", err)
		}

		//未找到设备认证方式信息，创建新的设备认证方式
		if err != nil {
			authMethod = user.AuthMethods{
				UserId:         newUserId,
				AuthType:       "device",
				AuthIdentifier: deviceInfo.Identifier,
				Verified:       true,
			}
			logger.Infof("create auth method: %v", authMethod)
			if err := tx.Create(&authMethod).Error; err != nil {
				l.Errorw("failed to create device auth method", logger.Field("new_user_id", newUserId),
					logger.Field("error", err.Error()))
				return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create device auth method failed: %v", err)
			}
		} else {
			//更新设备认证方式的用户ID为新用户ID
			authMethod.UserId = newUserId
			if err := tx.Save(&authMethod).Error; err != nil {
				l.Errorw("failed to update device auth method",
					logger.Field("identifier", deviceInfo.Identifier),
					logger.Field("error", err.Error()),
				)
				return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update device auth method failed: %v", err)
			}
		}

		//检查旧用户是否还有其他认证方式
		var count int64
		if err := tx.Model(&user.AuthMethods{}).Where("user_id = ?", oldUserId).Count(&count).Error; err != nil {
			l.Errorw("failed to query auth methods for old user",
				logger.Field("old_user_id", oldUserId),
				logger.Field("error", err.Error()),
			)
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query auth methods failed: %v", err)
		}

		//如果没有其他认证方式，禁用旧用户账号
		if count < 1 {
			//检查设备下是否有套餐，有套餐。就检查即将绑定过去的所有账户是否有套餐，如果有，那么检查两个套餐是否一致。如果一致就将即将删除的用户套餐，时间叠加到我绑定过去的用户套餐上面（如果套餐已过期就忽略）。新绑定设备的账户上套餐不一致或者不存在直接将套餐换绑即可
			var oldUserSubscribes []user.Subscribe
			err = tx.Where("user_id = ? AND status IN ?", oldUserId, []int64{0, 1}).Find(&oldUserSubscribes).Error
			if err != nil {
				l.Errorw("failed to query old user subscribes",
					logger.Field("old_user_id", oldUserId),
					logger.Field("error", err.Error()),
				)
				return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query old user subscribes failed: %v", err)
			}

			if len(oldUserSubscribes) > 0 {
				l.Infow("processing old user subscribes",
					logger.Field("old_user_id", oldUserId),
					logger.Field("new_user_id", newUserId),
					logger.Field("subscribe_count", len(oldUserSubscribes)),
				)

				for _, oldSub := range oldUserSubscribes {
					// 检查新用户是否有相同套餐ID的订阅
					var newUserSub user.Subscribe
					err = tx.Where("user_id = ? AND subscribe_id = ? AND status IN ?", newUserId, oldSub.SubscribeId, []int64{0, 1}).First(&newUserSub).Error

					if err != nil {
						// 新用户没有该套餐，直接换绑
						oldSub.UserId = newUserId
						if err := tx.Save(&oldSub).Error; err != nil {
							l.Errorw("failed to rebind subscribe to new user",
								logger.Field("subscribe_id", oldSub.Id),
								logger.Field("old_user_id", oldUserId),
								logger.Field("new_user_id", newUserId),
								logger.Field("error", err.Error()),
							)
							return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "rebind subscribe failed: %v", err)
						}
						l.Infow("rebind subscribe to new user",
							logger.Field("subscribe_id", oldSub.Id),
							logger.Field("new_user_id", newUserId),
						)
					} else {
						// 新用户已有该套餐，检查旧套餐是否过期
						now := time.Now()
						if oldSub.ExpireTime.After(now) {
							// 旧套餐未过期，叠加剩余时间
							remainingDuration := oldSub.ExpireTime.Sub(now)
							if newUserSub.ExpireTime.After(now) {
								// 新套餐未过期，叠加时间
								newUserSub.ExpireTime = newUserSub.ExpireTime.Add(remainingDuration)
							} else {
								newUserSub.ExpireTime = time.Now().Add(remainingDuration)
							}
							if err := tx.Save(&newUserSub).Error; err != nil {
								l.Errorw("failed to update subscribe expire time",
									logger.Field("subscribe_id", newUserSub.Id),
									logger.Field("error", err.Error()),
								)
								return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update subscribe expire time failed: %v", err)
							}
							l.Infow("merged subscribe time",
								logger.Field("subscribe_id", newUserSub.Id),
								logger.Field("new_expire_time", newUserSub.ExpireTime),
							)
						} else {
							l.Infow("old subscribe expired, skip merge",
								logger.Field("subscribe_id", oldSub.Id),
								logger.Field("expire_time", oldSub.ExpireTime),
							)
						}
						// 删除旧用户的套餐
						if err := tx.Delete(&oldSub).Error; err != nil {
							l.Errorw("failed to delete old subscribe",
								logger.Field("subscribe_id", oldSub.Id),
								logger.Field("error", err.Error()),
							)
							return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete old subscribe failed: %v", err)
						}
					}
				}
			}

			if err := tx.Model(&user.User{}).Where("id = ?", oldUserId).Delete(&user.User{}).Error; err != nil {
				l.Errorw("failed to disable old user",
					logger.Field("old_user_id", oldUserId),
					logger.Field("error", err.Error()),
				)
				return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "disable old user failed: %v", err)
			}
		}

		l.Infow("disabled old user (no other auth methods)",
			logger.Field("old_user_id", oldUserId),
		)

		// 更新设备绑定的用户id
		deviceInfo.UserId = newUserId
		deviceInfo.Ip = ip
		deviceInfo.UserAgent = userAgent
		deviceInfo.Enabled = true
		if err := tx.Save(deviceInfo).Error; err != nil {
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

	err = l.svcCtx.UserModel.ClearUserCache(l.ctx, users...)
	if err != nil {
		l.Errorw("failed to clear user cache after rebinding",
			logger.Field("old_user_id", oldUserId),
			logger.Field("new_user_id", newUserId),
			logger.Field("error", err.Error()),
		)
	}

	l.Infow("device rebound successfully",
		logger.Field("identifier", deviceInfo.Identifier),
		logger.Field("old_user_id", oldUserId),
		logger.Field("new_user_id", newUserId),
	)

	return nil
}
