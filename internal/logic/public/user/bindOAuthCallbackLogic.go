package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perfect-panel/server/pkg/constant"

	"github.com/perfect-panel/server/internal/model/auth"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/oauth/apple"
	"github.com/perfect-panel/server/pkg/oauth/google"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BindOAuthCallbackLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Bind OAuth Callback
func NewBindOAuthCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindOAuthCallbackLogic {
	return &BindOAuthCallbackLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type googleRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func (l *BindOAuthCallbackLogic) BindOAuthCallback(req *types.BindOAuthCallbackRequest) error {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	var err error
	switch req.Method {
	case "google":
		err = l.google(req)
	case "apple":
		err = l.apple(req)
	case "telegram":
		err = l.telegram(req)
	default:
		l.Errorw("oauth login method not support", logger.Field("method", req.Method))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "oauth login method not support: %v", req.Method)
	}
	if err != nil {
		l.Errorw("bind oauth callback failed: %v", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "bind oauth callback failed")
	}
	// update user info to redis
	err = l.svcCtx.UserModel.UpdateUserCache(l.ctx, u)
	if err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update user cache failed")
	}

	return nil
}
func (l *BindOAuthCallbackLogic) google(req *types.BindOAuthCallbackRequest) error {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	var request googleRequest
	err := tool.CloneMapToStruct(req.Callback.(map[string]interface{}), &request)
	if err != nil {
		l.Errorw("error CloneMapToStruct: %v", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "CloneMapToStruct failed")
	}
	// validate the state code
	redirect, err := l.svcCtx.Redis.Get(l.ctx, fmt.Sprintf("google:%s", request.State)).Result()
	if err != nil {
		l.Errorw("error get google state code: %v", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "get google state code failed")
	}
	// get google config
	authMethod, err := l.svcCtx.AuthModel.FindOneByMethod(l.ctx, "google")
	if err != nil {
		l.Errorw("error find google auth method: %v", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find google auth method failed")
	}
	var cfg auth.GoogleAuthConfig
	err = json.Unmarshal([]byte(authMethod.Config), &cfg)
	if err != nil {
		l.Errorw("error unmarshal google config: %v", logger.Field("config", authMethod.Config), logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "unmarshal google config failed")
	}
	client := google.New(&google.Config{
		ClientID:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  redirect,
	})
	token, err := client.Exchange(l.ctx, request.Code)
	if err != nil {
		l.Errorw("error exchange google token: %v", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "exchange google token failed")
	}
	googleUserInfo, err := client.GetUserInfo(token.AccessToken)
	if err != nil {
		l.Errorw("error get google user info: %v", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "get google user info failed")
	}
	// query user info
	userAuthMethod, err := l.svcCtx.UserModel.FindUserAuthMethodByOpenID(l.ctx, "google", googleUserInfo.OpenID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query user auth method failed")
	}
	if userAuthMethod.Id > 0 {
		return errors.Wrapf(xerr.NewErrCode(xerr.UserExist), "google user already exists")
	}
	// bind google
	userAuthMethod = &user.AuthMethods{
		UserId:         u.Id,
		AuthType:       "google",
		AuthIdentifier: googleUserInfo.OpenID,
		Verified:       true,
	}
	err = l.svcCtx.UserModel.InsertUserAuthMethods(l.ctx, userAuthMethod)
	if err != nil {
		l.Errorw("error insert user auth method: %v", logger.Field("error", err.Error()))
		return err
	}
	return nil
}

func (l *BindOAuthCallbackLogic) apple(req *types.BindOAuthCallbackRequest) error {
	// validate the state code
	_, err := l.svcCtx.Redis.Get(l.ctx, fmt.Sprintf("apple:%s", req.Callback.(map[string]interface{})["state"])).Result()
	if err != nil {
		l.Errorw("[BindOAuthCallbackLogic] Get State code error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "get apple state code failed: %v", err.Error())
	}
	appleAuth, err := l.svcCtx.AuthModel.FindOneByMethod(l.ctx, "apple")
	if err != nil {
		l.Errorw("[BindOAuthCallbackLogic] FindOneByMethod error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find apple auth method failed: %v", err.Error())
	}
	var appleCfg auth.AppleAuthConfig
	err = json.Unmarshal([]byte(appleAuth.Config), &appleCfg)
	if err != nil {
		l.Errorw("[BindOAuthCallbackLogic] Unmarshal error", logger.Field("error", err.Error()), logger.Field("config", appleAuth.Config))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "unmarshal apple config failed: %v", err.Error())
	}

	client, err := apple.New(apple.Config{
		ClientID:     appleCfg.ClientId,
		TeamID:       appleCfg.TeamID,
		KeyID:        appleCfg.KeyID,
		ClientSecret: appleCfg.ClientSecret,
		RedirectURI:  appleCfg.RedirectURL,
	})
	if err != nil {
		l.Errorw("[BindOAuthCallbackLogic] New apple client error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "new apple client failed: %v", err.Error())
	}
	// verify web token
	resp, err := client.VerifyWebToken(l.ctx, req.Callback.(map[string]interface{})["code"].(string))
	if err != nil {
		l.Errorw("[BindOAuthCallbackLogic] VerifyWebToken error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "verify web token failed: %v", err.Error())
	}
	if resp.Error != "" {
		l.Errorw("[BindOAuthCallbackLogic] VerifyWebToken error", logger.Field("error", resp.Error))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "verify web token failed: %v", resp.Error)
	}
	// query apple user unique id
	appleUnique, err := apple.GetUniqueID(resp.IDToken)
	if err != nil {
		l.Errorw("[BindOAuthCallbackLogic] GetUniqueID error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "get apple unique id failed: %v", err.Error())
	}
	// query user by apple unique id
	userAuthMethod, err := l.svcCtx.UserModel.FindUserAuthMethodByOpenID(l.ctx, "apple", appleUnique)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorw("[BindOAuthCallbackLogic] FindUserAuthMethodByOpenID error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find user auth method by openid failed: %v", err.Error())
	}
	if userAuthMethod.Id > 0 {
		l.Errorw("[BindOAuthCallbackLogic] User already exists")
		return errors.Wrapf(xerr.NewErrCode(xerr.UserExist), "apple user already exists")
	}
	// query user info
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	// bind apple
	userAuthMethod = &user.AuthMethods{
		UserId:         u.Id,
		AuthType:       "apple",
		AuthIdentifier: appleUnique,
		Verified:       true,
	}
	err = l.svcCtx.UserModel.InsertUserAuthMethods(l.ctx, userAuthMethod)
	if err != nil {
		l.Errorw("[BindOAuthCallbackLogic] InsertUserAuthMethods error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "insert user auth method failed: %v", err.Error())
	}
	return nil
}

func (l *BindOAuthCallbackLogic) telegram(req *types.BindOAuthCallbackRequest) error {
	return nil
}
