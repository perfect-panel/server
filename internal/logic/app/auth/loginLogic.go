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

type LoginLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

// Login
func NewLoginLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.AppAuthRequest) (resp *types.AppAuthRespone, err error) {

	loginStatus := false
	var userInfo *user.User
	// Record login status
	defer func(svcCtx *svc.ServiceContext) {
		if userInfo != nil && userInfo.Id != 0 {
			if err := svcCtx.UserModel.InsertLoginLog(l.ctx, &user.LoginLog{
				UserId:    userInfo.Id,
				LoginIP:   l.ctx.ClientIP(),
				UserAgent: l.ctx.Request.UserAgent(),
				Success:   &loginStatus,
			}); err != nil {
				l.Errorw("InsertLoginLog Error", logger.Field("error", err.Error()))
			}
		}
	}(l.svcCtx)

	resp = &types.AppAuthRespone{}
	//query user
	userInfo, err = findUserByMethod(l.ctx, l.svcCtx, req.Method, req.Identifier, req.Account, req.AreaCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.UserPasswordError), "user password")
		}
		return resp, err
	}

	switch req.Method {
	case authmethod.Email:

		if !l.svcCtx.Config.Email.Enable {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.EmailNotEnabled), "Email function is not enabled yet")
		}

		if req.Code != "" {
			cacheKey := fmt.Sprintf("%s:%s:%s", config.AuthCodeCacheKey, constant.Security.String(), req.Account)
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
			l.svcCtx.Redis.Del(l.ctx, cacheKey)
		} else {
			// Verify password
			if !tool.VerifyPassWord(req.Password, userInfo.Password) {
				return nil, errors.Wrapf(xerr.NewErrCode(xerr.UserPasswordError), "user password")
			}
		}
	case authmethod.Mobile:
		if !l.svcCtx.Config.Mobile.Enable {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.SmsNotEnabled), "sms login is not enabled")
		}
		phoneNumber, err := phone.FormatToE164(req.AreaCode, req.Account)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.TelephoneError), "Invalid phone number")
		}

		if req.Code != "" {
			cacheKey := fmt.Sprintf("%s:%s:%s", config.AuthCodeTelephoneCacheKey, constant.Security, phoneNumber)
			value, err := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
			if err != nil {
				l.Errorw("Redis Error", logger.Field("error", err.Error()), logger.Field("cacheKey", cacheKey))
				return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
			}

			if value == "" {
				return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
			}

			var payload common.CacheKeyPayload
			if err := json.Unmarshal([]byte(value), &payload); err != nil {
				l.Errorw("[SendSmsCode]: Unmarshal Error", logger.Field("error", err.Error()), logger.Field("value", value))
			}
			if payload.Code != req.Code {
				return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
			}
			l.svcCtx.Redis.Del(l.ctx, cacheKey)
		} else {
			// Verify password
			if !tool.VerifyPassWord(req.Password, userInfo.Password) {
				return nil, errors.Wrapf(xerr.NewErrCode(xerr.UserPasswordError), "user password")
			}
		}
	case authmethod.Device:
	default:
		return nil, existError(req.Method)
	}

	device, err := l.svcCtx.UserModel.FindOneDeviceByIdentifier(l.ctx, req.Identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if req.Method == authmethod.Device {
				return nil, errors.Wrapf(xerr.NewErrCode(xerr.UserNotExist), "device not exist")
			}
			//Add User Device
			userInfo.UserDevices = append(userInfo.UserDevices, user.Device{
				UserAgent:  req.UserAgent,
				Identifier: req.Identifier,
				Ip:         l.ctx.ClientIP(),
			})
			err = l.svcCtx.UserModel.Update(l.ctx, userInfo)
			if err != nil {
				l.Errorw("[UpdateUserBindDevice] Fail", logger.Field("error", err.Error()))
			}
		}
	} else {
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
	if err = l.svcCtx.Redis.Set(l.ctx, sessionIdCacheKey, userInfo.Id, time.Duration(l.svcCtx.Config.JwtAuth.AccessExpire)*time.Second).Err(); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "set session id error: %v", err.Error())
	}

	resp.Token = token
	return
}
