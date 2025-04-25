package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/perfect-panel/ppanel-server/pkg/authmethod"

	"github.com/gin-gonic/gin"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/internal/logic/common"
	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/constant"
	"github.com/perfect-panel/ppanel-server/pkg/jwt"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/phone"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/uuidx"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CacheKeyPayload struct {
	Code   string `json:"code"`
	LastAt int64  `json:"lastAt"`
}
type RegisterLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

// Register
func NewRegisterLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.AppAuthRequest) (resp *types.AppAuthRespone, err error) {
	resp = &types.AppAuthRespone{}
	var referer *user.User
	c := l.svcCtx.Config.Register
	// Check if the registration is stopped
	if c.StopRegister {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.StopRegister), "stop register")
	}

	if req.Invite == "" {
		if l.svcCtx.Config.Invite.ForcedInvite {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.InviteCodeError), "invite code is required")
		}
	} else {
		// Check if the invite code is valid
		referer, err = l.svcCtx.UserModel.FindOneByReferCode(l.ctx, req.Invite)
		if err != nil {
			l.Errorw("FindOneByReferCode Error", logger.Field("error", err))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.InviteCodeError), "invite code is invalid")
		}
	}

	if req.Password == "" {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.PasswordIsEmpty), "Password  required")
	}

	userInfo, err := findUserByMethod(l.ctx, l.svcCtx, req.Method, req.Identifier, req.Account, req.AreaCode)
	if err == nil && userInfo != nil {
		return nil, existError(req.Method)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// Generate password
	pwd := tool.EncodePassWord(req.Password)
	userInfo = &user.User{
		Password: pwd,
	}
	if referer != nil {
		userInfo.RefererId = referer.Id
	}
	switch req.Method {
	case authmethod.Email:
		if !l.svcCtx.Config.Email.Enable {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.EmailNotEnabled), "Email function is not enabled yet")
		}
		if l.svcCtx.Config.Email.EnableVerify {
			cacheKey := fmt.Sprintf("%s:%s:%s", config.AuthCodeCacheKey, constant.Register.String(), req.Account)
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
		}
		userInfo.AuthMethods = []user.AuthMethods{{
			AuthType:       authmethod.Email,
			AuthIdentifier: req.Account,
		}}

	case authmethod.Mobile:
		if req.AreaCode == "" {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.TelephoneAreaCodeIsEmpty), "area code required")
		}

		if !l.svcCtx.Config.Mobile.Enable {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.SmsNotEnabled), "sms login is not enabled")
		}
		phoneNumber, err := phone.FormatToE164(req.AreaCode, req.Account)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.TelephoneError), "Invalid phone number")
		}
		cacheKey := fmt.Sprintf("%s:%s:%s", config.AuthCodeTelephoneCacheKey, constant.Register, phoneNumber)
		value, err := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
		if err != nil {
			l.Errorw("Redis Error", logger.Field("error", err.Error()), logger.Field("cacheKey", cacheKey))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}

		var payload CacheKeyPayload
		_ = json.Unmarshal([]byte(value), &payload)
		if payload.Code != req.Code {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}
		userInfo.AuthMethods = []user.AuthMethods{{
			AuthType:       authmethod.Mobile,
			AuthIdentifier: phoneNumber,
			Verified:       true,
		}}
	case authmethod.Device:
		oneDevice, err := l.svcCtx.UserModel.FindOneDeviceByIdentifier(l.ctx, req.Identifier)
		if err == nil && oneDevice != nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DeviceExist), "device exist")
		}
	default:
		return nil, existError(req.Method)
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
		//Delete Other User Device
		err = l.svcCtx.UserModel.DeleteDevice(l.ctx, device.Id)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "delete old user device failed: %v", err.Error())
		} else {
			//User Add Device
			userInfo.UserDevices = append(userInfo.UserDevices, user.Device{
				Ip:         l.ctx.ClientIP(),
				Identifier: req.Identifier,
				UserAgent:  req.UserAgent,
			})
		}
	}

	err = l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		// Save user information
		if err := db.Create(userInfo).Error; err != nil {
			return err
		}
		// Generate ReferCode
		userInfo.ReferCode = uuidx.UserInviteCode(userInfo.Id)
		// Update ReferCode
		if err := db.Model(&user.User{}).Where("id = ?", userInfo.Id).Update("refer_code", userInfo.ReferCode).Error; err != nil {
			return err
		}
		if l.svcCtx.Config.Register.EnableTrial {
			// Active trial
			if err = l.activeTrial(userInfo.Id); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "insert user info failed: %v", err.Error())
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

func (l *RegisterLogic) activeTrial(uid int64) error {
	sub, err := l.svcCtx.SubscribeModel.FindOne(l.ctx, l.svcCtx.Config.Register.TrialSubscribe)
	if err != nil {
		return err
	}
	userSub := &user.Subscribe{
		Id:          0,
		UserId:      uid,
		OrderId:     0,
		SubscribeId: sub.Id,
		StartTime:   time.Now(),
		ExpireTime:  tool.AddTime(l.svcCtx.Config.Register.TrialTimeUnit, l.svcCtx.Config.Register.TrialTime, time.Now()),
		Traffic:     sub.Traffic,
		Download:    0,
		Upload:      0,
		Token:       uuidx.SubscribeToken(fmt.Sprintf("Trial-%v", uid)),
		UUID:        uuidx.NewUUID().String(),
		Status:      1,
	}
	return l.svcCtx.UserModel.InsertSubscribe(l.ctx, userSub)
}
