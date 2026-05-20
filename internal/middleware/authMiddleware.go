package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/jwt"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/result"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

func AuthMiddleware(svc *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, err := AuthenticateRequest(c.Request.Context(), svc, c.GetHeader("Authorization"), c.Request.URL.Path)
		if err != nil {
			result.HttpResult(c, nil, err)
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func AuthenticateRequest(ctx context.Context, svc *svc.ServiceContext, token string, path string) (context.Context, error) {
	jwtConfig := svc.Config.JwtAuth
	if token == "" {
		logger.WithContext(ctx).Debug("[AuthMiddleware] Token Empty")
		return ctx, errors.Wrapf(xerr.NewErrCode(xerr.ErrorTokenEmpty), "Token Empty")
	}

	claims, err := jwt.ParseJwtToken(token, jwtConfig.AccessSecret)
	if err != nil {
		logger.WithContext(ctx).Debug("[AuthMiddleware] ParseJwtToken", logger.Field("error", err.Error()), logger.Field("token", token))
		return ctx, errors.Wrapf(xerr.NewErrCode(xerr.ErrorTokenExpire), "Token Invalid")
	}

	loginType := ""
	if claims["LoginType"] != nil {
		loginType = claims["LoginType"].(string)
	}

	userId := int64(claims["UserId"].(float64))
	sessionId := claims["SessionId"].(string)
	sessionIdCacheKey := fmt.Sprintf("%v:%v", config.SessionIdKey, sessionId)
	value, err := svc.Redis.Get(ctx, sessionIdCacheKey).Result()
	if err != nil {
		logger.WithContext(ctx).Debug("[AuthMiddleware] Redis Get", logger.Field("error", err.Error()), logger.Field("sessionId", sessionId))
		return ctx, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}

	if value != fmt.Sprintf("%v", userId) {
		logger.WithContext(ctx).Debug("[AuthMiddleware] Invalid Access", logger.Field("userId", userId), logger.Field("sessionId", sessionId))
		return ctx, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}

	userInfo, err := svc.Store.User().FindOne(ctx, userId)
	if err != nil {
		logger.WithContext(ctx).Debug("[AuthMiddleware] UserModel FindOne", logger.Field("error", err.Error()), logger.Field("userId", userId))
		return ctx, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Database Query Error")
	}

	paths := strings.Split(path, "/")
	if tool.StringSliceContains(paths, "admin") && !*userInfo.IsAdmin {
		logger.WithContext(ctx).Debug("[AuthMiddleware] Not Admin User", logger.Field("userId", userId), logger.Field("sessionId", sessionId))
		return ctx, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}

	ctx = context.WithValue(ctx, constant.LoginType, loginType)
	ctx = context.WithValue(ctx, constant.CtxKeyUser, userInfo)
	ctx = context.WithValue(ctx, constant.CtxKeySessionID, sessionId)
	return ctx, nil
}
