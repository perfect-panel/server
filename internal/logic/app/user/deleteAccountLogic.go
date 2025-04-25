package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/internal/logic/common"
	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/constant"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type DeleteAccountLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete Account
func NewDeleteAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAccountLogic {
	return &DeleteAccountLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAccountLogic) DeleteAccount(req *types.DeleteAccountRequest) error {
	userInfo, exists := l.ctx.Value(constant.CtxKeyUser).(user.User)
	if !exists {
		return nil
	}

	var account string
	for _, authMethod := range userInfo.AuthMethods {
		if authMethod.AuthType == req.Method {
			account = authMethod.AuthIdentifier
			break
		}
	}
	if account == "" {
		return nil
	}

	if req.Method == "email" {
		emailConfig := l.svcCtx.Config.Email

		if !emailConfig.Enable {
			return errors.Wrapf(xerr.NewErrCode(xerr.EmailNotEnabled), "Email function is not enabled yet")
		}

		if emailConfig.EnableVerify {
			cacheKey := fmt.Sprintf("%s:%s:%s", config.AuthCodeCacheKey, constant.Security, account)
			value, err := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
			if err != nil {
				l.Errorw("Redis Error", logger.Field("error", err.Error()), logger.Field("cacheKey", cacheKey))
				return errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
			}

			var payload common.CacheKeyPayload
			err = json.Unmarshal([]byte(value), &payload)
			if err != nil {
				l.Errorw("Unmarshal Error", logger.Field("error", err.Error()), logger.Field("value", value))
				return errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
			}

			if payload.Code != req.Code {
				return errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
			}
		}
	} else {
		cacheKey := fmt.Sprintf("%s:%s:%s", config.AuthCodeTelephoneCacheKey, constant.Security, account)
		value, err := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
		if err != nil {
			l.Errorw("Redis Error", logger.Field("error", err.Error()), logger.Field("cacheKey", cacheKey))
			return errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}

		if value == "" {
			return errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}

		var payload common.CacheKeyPayload
		if err := json.Unmarshal([]byte(value), &payload); err != nil {
			l.Errorw("[SendSmsCode]: Unmarshal Error", logger.Field("error", err.Error()), logger.Field("value", value))
		}
		if payload.Code != req.Code {
			return errors.Wrapf(xerr.NewErrCode(xerr.VerifyCodeError), "code error")
		}
	}
	err := l.svcCtx.UserModel.Delete(l.ctx, userInfo.Id)
	if err != nil {
		l.Errorw("update user password error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update user password")
	}
	return nil
}
