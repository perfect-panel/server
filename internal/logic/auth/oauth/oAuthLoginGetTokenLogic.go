package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/internal/model/auth"
	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/jwt"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/oauth/apple"
	"github.com/perfect-panel/ppanel-server/pkg/oauth/google"
	"github.com/perfect-panel/ppanel-server/pkg/oauth/telegram"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/uuidx"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type googleRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}
type OAuthLoginGetTokenLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOAuthLoginGetTokenLogic OAuth login get token
func NewOAuthLoginGetTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OAuthLoginGetTokenLogic {
	return &OAuthLoginGetTokenLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OAuthLoginGetTokenLogic) OAuthLoginGetToken(req *types.OAuthLoginGetTokenRequest, ip, userAgent string) (resp *types.LoginResponse, err error) {
	loginStatus := false
	var userInfo *user.User
	// Record login status
	defer func(svcCtx *svc.ServiceContext) {
		if userInfo != nil && userInfo.Id != 0 {
			if err := svcCtx.UserModel.InsertLoginLog(l.ctx, &user.LoginLog{
				UserId:    userInfo.Id,
				LoginIP:   ip,
				UserAgent: userAgent,
				Success:   &loginStatus,
			}); err != nil {
				l.Errorw("error insert login log: %v", logger.Field("error", err.Error()))
			}
		}
	}(l.svcCtx)
	switch req.Method {
	case "google":
		userInfo, err = l.google(req)
	case "apple":
		userInfo, err = l.apple(req)
	case "telegram":
		userInfo, err = l.telegram(req)
	default:
		l.Errorw("oauth login method not support: %v", logger.Field("method", req.Method))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "oauth login method not support: %v", req.Method)
	}
	if err != nil {
		return nil, err
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
	loginStatus = true
	return &types.LoginResponse{
		Token: token,
	}, nil
}

func (l *OAuthLoginGetTokenLogic) google(req *types.OAuthLoginGetTokenRequest) (*user.User, error) {
	var request googleRequest
	err := tool.CloneMapToStruct(req.Callback.(map[string]interface{}), &request)
	if err != nil {
		l.Errorw("error CloneMapToStruct: %v", logger.Field("error", err.Error()))
		return nil, err
	}
	// validate the state code
	redirect, err := l.svcCtx.Redis.Get(l.ctx, fmt.Sprintf("google:%s", request.State)).Result()
	if err != nil {
		l.Errorw("error get google state code: %v", logger.Field("error", err.Error()))
		return nil, err
	}
	// get google config
	authMethod, err := l.svcCtx.AuthModel.FindOneByMethod(l.ctx, "google")
	if err != nil {
		l.Errorw("error find google auth method: %v", logger.Field("error", err.Error()))
		return nil, err
	}
	var cfg auth.GoogleAuthConfig
	err = cfg.Unmarshal(authMethod.Config)
	if err != nil {
		l.Errorw("error unmarshal google config: %v", logger.Field("config", authMethod.Config), logger.Field("error", err.Error()))
		return nil, err
	}
	client := google.New(&google.Config{
		ClientID:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  redirect,
	})
	token, err := client.Exchange(l.ctx, request.Code)
	if err != nil {
		l.Errorw("error exchange google token: %v", logger.Field("error", err.Error()))
		return nil, err
	}
	googleUserInfo, err := client.GetUserInfo(token.AccessToken)
	if err != nil {
		l.Errorw("error get google user info: %v", logger.Field("error", err.Error()))
		return nil, err
	}
	// query user info
	userAuthMethod, err := l.svcCtx.UserModel.FindUserAuthMethodByOpenID(l.ctx, "google", googleUserInfo.OpenID)
	if err != nil {
		if errors.As(err, &gorm.ErrRecordNotFound) {
			return l.register(googleUserInfo.Email, googleUserInfo.Picture, "google", googleUserInfo.OpenID)
		}
		return nil, err
	}
	return l.svcCtx.UserModel.FindOne(l.ctx, userAuthMethod.UserId)
}

func (l *OAuthLoginGetTokenLogic) apple(req *types.OAuthLoginGetTokenRequest) (*user.User, error) {
	// validate the state code
	_, err := l.svcCtx.Redis.Get(l.ctx, fmt.Sprintf("apple:%s", req.Callback.(map[string]interface{})["state"])).Result()
	if err != nil {
		l.Errorw("[AppleLoginCallback] Get State code error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "get apple state code failed: %v", err.Error())
	}
	appleAuth, err := l.svcCtx.AuthModel.FindOneByMethod(l.ctx, "apple")
	if err != nil {
		l.Errorw("[AppleLoginCallback] FindOneByMethod error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find apple auth method failed: %v", err.Error())
	}
	var appleCfg auth.AppleAuthConfig
	err = appleCfg.Unmarshal(appleAuth.Config)
	if err != nil {
		l.Errorw("[AppleLoginCallback] Unmarshal error", logger.Field("error", err.Error()), logger.Field("config", appleAuth.Config))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "unmarshal apple config failed: %v", err.Error())
	}

	client, err := apple.New(apple.Config{
		ClientID:     appleCfg.ClientId,
		TeamID:       appleCfg.TeamID,
		KeyID:        appleCfg.KeyID,
		ClientSecret: appleCfg.ClientSecret,
		RedirectURI:  appleCfg.RedirectURL,
	})
	if err != nil {
		l.Errorw("[AppleLoginCallback] New apple client error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "new apple client failed: %v", err.Error())
	}
	// verify web token
	resp, err := client.VerifyWebToken(l.ctx, req.Callback.(map[string]interface{})["code"].(string))
	if err != nil {
		l.Errorw("[AppleLoginCallback] VerifyWebToken error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "verify web token failed: %v", err.Error())
	}
	if resp.Error != "" {
		l.Errorw("[AppleLoginCallback] VerifyWebToken error", logger.Field("error", resp.Error))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "verify web token failed: %v", resp.Error)
	}
	// query apple user unique id
	appleUnique, err := apple.GetUniqueID(resp.IDToken)
	if err != nil {
		l.Errorw("[AppleLoginCallback] GetUniqueID error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "get apple unique id failed: %v", err.Error())
	}
	// get apple user info
	appleUserInfo, err := apple.GetClaims(resp.AccessToken)
	if err != nil {
		l.Errorw("[AppleLoginCallback] GetClaims error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "get apple user info failed: %v", err.Error())
	}
	// query user by apple unique id
	userAuthMethod, err := l.svcCtx.UserModel.FindUserAuthMethodByOpenID(l.ctx, "apple", appleUnique)
	if err != nil {
		// if user not exist, handle register
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return l.register((*appleUserInfo)["email"].(string), "", "apple", appleUnique)
		}
		l.Errorw("[AppleLoginCallback] FindUserAuthMethodByOpenID error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find user auth method by openid failed: %v", err.Error())
	}
	// query user info
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, userAuthMethod.UserId)

	if err != nil {
		l.Errorw(
			"[AppleLoginCallback] FindOne error",
			logger.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find user info failed: %v", err.Error())
	}

	return userInfo, nil
}

func (l *OAuthLoginGetTokenLogic) telegram(req *types.OAuthLoginGetTokenRequest) (*user.User, error) {
	appleAuth, err := l.svcCtx.AuthModel.FindOneByMethod(l.ctx, "telegram")
	if err != nil {
		l.Errorw("[OAuthLoginGetToken] FindOneByMethod error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find telegram auth method failed: %v", err.Error())
	}
	var telegramCfg auth.TelegramAuthConfig
	err = json.Unmarshal([]byte(appleAuth.Config), &telegramCfg)
	if err != nil {
		l.Errorw("[OAuthLoginGetToken] Unmarshal error", logger.Field("error", err.Error()), logger.Field("config", appleAuth.Config))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "unmarshal telegram config failed: %v", err.Error())
	}
	encodeText := req.Callback.(map[string]interface{})["tgAuthResult"].(string)
	// base64 decode
	callbackData, err := telegram.ParseAndValidateBase64([]byte(encodeText), telegramCfg.BotToken)
	if err != nil {
		l.Errorw("[TelegramLoginCallback] ParseAndValidateBase64 error", logger.Field("error", err.Error()))
		return nil, err
	}
	// 验证数据有效期
	if time.Now().Unix()-*callbackData.AuthDate > 86400 {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "auth date expired")
	}
	// query user auth info
	userAuthMethod, err := l.svcCtx.UserModel.FindUserAuthMethodByOpenID(l.ctx, "telegram", fmt.Sprintf("%v", *callbackData.Id))
	if err != nil {
		if errors.As(err, &gorm.ErrRecordNotFound) {
			return l.register(fmt.Sprintf("%v@%s", *callbackData.Id, "qq.com"), *callbackData.PhotoUrl, "telegram", fmt.Sprintf("%v", callbackData.Id))
		}
	}
	// query user info
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, userAuthMethod.UserId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find user info failed: %v", err.Error())
	}
	return userInfo, nil
}

func (l *OAuthLoginGetTokenLogic) register(email, avatar, method, openid string) (*user.User, error) {
	if l.svcCtx.Config.Invite.ForcedInvite {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InviteCodeError), "invite code is required")
	}
	var userInfo *user.User
	err := l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		err := db.Model(&user.User{}).Where("email = ?", email).First(&userInfo).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if userInfo.Id != 0 {
			return errors.Wrapf(xerr.NewErrCode(xerr.UserExist), "user email exist: %v", email)
		}
		userInfo = &user.User{
			Avatar: avatar,
		}
		if err := db.Create(userInfo).Error; err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create user info failed: %v", err.Error())
		}
		// Generate ReferCode
		userInfo.ReferCode = uuidx.UserInviteCode(userInfo.Id)
		// Update ReferCode
		err = db.Where("id = ?", userInfo.Id).Update("refer_code", userInfo.ReferCode).Error
		if err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update refer code failed: %v", err.Error())
		}
		authMethod := &user.AuthMethods{
			UserId:         userInfo.Id,
			AuthType:       method,
			AuthIdentifier: openid,
			Verified:       true,
		}
		if err = db.Create(authMethod).Error; err != nil {
			l.Errorw("error create auth method: %v", logger.Field("error", err.Error()))
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create auth method failed: %v", err.Error())
		}
		if email != "" {
			authMethod = &user.AuthMethods{
				UserId:         userInfo.Id,
				AuthType:       "email",
				AuthIdentifier: email,
				Verified:       true,
			}
			if err := db.Create(authMethod).Error; err != nil {
				l.Errorw("error create auth method: %v", logger.Field("error", err.Error()))
				return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create auth method failed: %v", err.Error())
			}
		}
		if l.svcCtx.Config.Register.EnableTrial {
			// Active trial
			if err = l.activeTrial(userInfo.Id); err != nil {
				return err
			}
		}
		return nil
	})
	return userInfo, err
}

func (l *OAuthLoginGetTokenLogic) activeTrial(uid int64) error {
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
