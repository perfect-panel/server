package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/perfect-panel/ppanel-server/pkg/authmethod"

	"github.com/gin-gonic/gin"

	"github.com/perfect-panel/ppanel-server/pkg/constant"
	"github.com/perfect-panel/ppanel-server/pkg/phone"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/internal/logic/common"
	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/jwt"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/uuidx"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type ResetPasswordLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

// Reset Password
func NewResetPasswordLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(req *types.AppAuthRequest) (resp *types.AppAuthRespone, err error) {
	resp = &types.AppAuthRespone{}
	userInfo, err := findUserByMethod(l.ctx, l.svcCtx, req.Method, req.Identifier, req.Account, req.AreaCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.UserNotExist), "query user info failed")
		}
		l.Errorw("FindOneByEmail Error", logger.Field("error", err))
		return nil, err
	}

	switch req.Method {
	case authmethod.Mobile:
		if !l.svcCtx.Config.Mobile.Enable {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.SmsNotEnabled), "sms login is not enabled")
		}
		phoneNumber, err := phone.FormatToE164(req.AreaCode, req.Account)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.TelephoneError), "Invalid phone number")
		}
		cacheKey := fmt.Sprintf("%s:%s:%s", config.AuthCodeTelephoneCacheKey, constant.Security, phoneNumber)
		value, err := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
		if err != nil {
			l.Errorw("Redis Error", logger.Field("error", err.Error()), logger.Field("cacheKey", cacheKey))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}

		var payload common.CacheKeyPayload
		err = json.Unmarshal([]byte(value), &payload)
		if err != nil {
			l.Errorw("Unmarshal Error", logger.Field("error", err.Error()), logger.Field("value", value))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}

		if payload.Code != req.Code {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}
	case authmethod.Email:
		if !l.svcCtx.Config.Email.Enable {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.EmailNotEnabled), "Email function is not enabled yet")
		}

		cacheKey := fmt.Sprintf("%s:%s:%s", config.AuthCodeCacheKey, constant.Security.String(), req.Account)
		value, err := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
		if err != nil {
			l.Errorw("Redis Error", logger.Field("error", err.Error()), logger.Field("cacheKey", cacheKey))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}

		var payload CacheKeyPayload
		err = json.Unmarshal([]byte(value), &payload)
		if err != nil {
			l.Errorw("Unmarshal Error", logger.Field("error", err.Error()), logger.Field("value", value))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}

		if payload.Code != req.Code {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}
	default:
		return nil, errors.New("unknown method")
	}

	userInfo.Password = tool.EncodePassWord(req.Password)
	err = l.svcCtx.UserModel.Update(l.ctx, userInfo)
	if err != nil {
		l.Errorw("UpdateUser Error", logger.Field("error", err))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update user info failed: %v", err.Error())
	}

	device, err := l.svcCtx.UserModel.FindOneDeviceByIdentifier(l.ctx, req.Identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//Add User Device
			userInfo.UserDevices = append(userInfo.UserDevices, user.Device{
				Ip:         l.ctx.ClientIP(),
				Identifier: req.Identifier,
				UserAgent:  req.UserAgent,
			})
		} else {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query user info failed: %v", err.Error())
		}
	} else {
		if device.UserId != userInfo.Id {
			//Change the user who owns the device
			if device.UserId != userInfo.Id {
				device.UserId = userInfo.Id
			}
			device.Ip = l.ctx.ClientIP()
			err = l.svcCtx.UserModel.UpdateDevice(l.ctx, device)
			if err != nil {
				l.Errorw("[UpdateUserBindDevice] Fail", logger.Field("error", err.Error()))
			}
		}
	}

	// Generate session id
	sessionId := uuidx.NewUUID().String()
	// Generate token
	token, err := jwt.NewJwtToken(
		l.svcCtx.Config.JwtAuth.AccessSecret,
		time.Now().Unix(),
		l.svcCtx.Config.JwtAuth.AccessExpire,
		jwt.WithOption("UserId", userInfo.Id),
		jwt.WithOption("SessionId", sessionId),
	)
	if err != nil {
		l.Logger.Error("[UserLogin] token generate error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "token generate error: %v", err.Error())
	}

	sessionIdCacheKey := fmt.Sprintf("%v:%v", config.SessionIdKey, sessionId)
	if err := l.svcCtx.Redis.Set(l.ctx, sessionIdCacheKey, userInfo.Id, time.Duration(l.svcCtx.Config.JwtAuth.AccessExpire)*time.Second).Err(); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "set session id error: %v", err.Error())
	}
	resp.Token = token
	return
}
