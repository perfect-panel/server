package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	pluginv1 "github.com/perfect-panel/server/api/plugin/v1"
	"github.com/perfect-panel/server/internal/middleware"
	usermodel "github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/plugin"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/result"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

// RegisterPluginHandlers 注册固定插件入口，具体插件路由由 Manager 动态分发。
func RegisterPluginHandlers(router *hertzx.Engine, svcCtx *svc.ServiceContext, mgr *plugin.Manager) {
	handler := buildPluginDispatcher(svcCtx, mgr)
	router.Any("/v1/plugin/:plugin", handler)
	router.Any("/v1/plugin/:plugin/*path", handler)
	logger.Info("registered plugin dispatcher")
}

func buildPluginDispatcher(svcCtx *svc.ServiceContext, mgr *plugin.Manager) hertzx.HandlerFunc {
	return func(c *hertzx.Context) {
		pluginName := c.Param("plugin")
		pluginPath := normalizePluginDispatchPath(c.Param("path"))

		route, ok := mgr.FindRoute(pluginName, c.Request.Method, pluginPath)
		if !ok {
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": fmt.Sprintf("plugin route not found: %s %s%s", c.Request.Method, pluginName, pluginPath),
			})
			return
		}

		after, ok := applyPluginRouteMiddleware(c, svcCtx, mgr, route)
		for i := len(after) - 1; i >= 0; i-- {
			defer after[i]()
		}
		if !ok {
			return
		}

		timeout := mgr.RequestTimeout()
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		instance := mgr.GetPlugin(pluginName)
		if instance == nil || instance.Pool == nil {
			c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"error": fmt.Sprintf("plugin %q is not ready", pluginName),
			})
			return
		}

		req := buildPluginHandleRequest(c)
		resp, err := mgr.CallPlugin(ctx, pluginName, route.Handler, req)
		if err != nil {
			logger.Errorf("plugin %q handler %q error: %v", pluginName, route.Handler, err)
			result.HttpResult(c, nil, fmt.Errorf("plugin error: %w", err))
			return
		}

		writePluginResponse(c, resp)
	}
}

func normalizePluginDispatchPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "*" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
		if path == "" {
			return "/"
		}
	}
	return path
}

func applyPluginRouteMiddleware(c *hertzx.Context, svcCtx *svc.ServiceContext, mgr *plugin.Manager, route plugin.RouteRegistration) ([]func(), bool) {
	after := make([]func(), 0)
	for _, mwName := range route.Middleware {
		switch mwName {
		case "auth":
			if !applyAuthMiddleware(c, svcCtx) {
				return after, false
			}
		case "device":
			deviceAfter, ok := applyDeviceMiddleware(c, svcCtx)
			if !ok {
				return after, false
			}
			if deviceAfter != nil {
				after = append(after, deviceAfter)
			}
		default:
			mw, ok := mgr.FindMiddleware(route.PluginName, mwName)
			if !ok {
				c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error": fmt.Sprintf("plugin middleware %q not found", mwName),
				})
				return after, false
			}
			if !applyWASMMiddleware(c, mgr, mw) {
				return after, false
			}
		}
	}
	return after, true
}

func applyAuthMiddleware(c *hertzx.Context, svcCtx *svc.ServiceContext) bool {
	ctx, err := middleware.AuthenticateRequest(c.Request.Context(), svcCtx, c.GetHeader("Authorization"), c.Request.URL.Path)
	if err != nil {
		result.HttpResult(c, nil, err)
		c.Abort()
		return false
	}
	c.Request = c.Request.WithContext(ctx)
	return true
}

func applyDeviceMiddleware(c *hertzx.Context, svcCtx *svc.ServiceContext) (func(), bool) {
	if !svcCtx.Config.Device.Enable {
		return nil, true
	}
	if svcCtx.Config.Device.SecuritySecret == "" {
		result.HttpResult(c, nil, errors.Wrapf(xerr.NewErrCode(xerr.SecretIsEmpty), "Secret is empty"))
		c.Abort()
		return nil, false
	}

	ctx := c.Request.Context()
	if ctx.Value(constant.CtxKeyUser) == nil && c.GetHeader("Login-Type") != "" {
		ctx = context.WithValue(ctx, constant.LoginType, c.GetHeader("Login-Type"))
		c.Request = c.Request.WithContext(ctx)
	}
	loginType, ok := c.Request.Context().Value(constant.LoginType).(string)
	if !ok || loginType != "device" {
		return nil, true
	}

	rw := middleware.NewResponseWriter(c, svcCtx)
	if !rw.Decrypt() {
		result.HttpResult(c, nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidCiphertext), "Invalid ciphertext"))
		c.Abort()
		return nil, false
	}
	hertzx.SyncRequestURI(c)
	hertzx.SyncRequestBody(c)
	c.Writer = rw
	return rw.FlushAbort, true
}

func applyWASMMiddleware(c *hertzx.Context, mgr *plugin.Manager, mw plugin.MiddlewareRegistration) bool {
	timeout := mgr.RequestTimeout()
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	req := buildPluginHandleRequest(c)
	resp, err := mgr.CallPluginMiddleware(ctx, mw.PluginName, mw.Handler, req)
	if err != nil {
		logger.Errorf("plugin %q middleware %q error: %v", mw.PluginName, mw.Handler, err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
		c.Abort()
		return false
	}
	if resp == nil {
		return true
	}

	for k, v := range resp.Headers {
		if resp.Action == "modify" {
			c.Request.Header.Set(k, v)
		} else {
			c.Header(k, v)
		}
	}

	if resp.Action == "abort" {
		if len(resp.Body) > 0 {
			c.Writer.WriteHeader(int(resp.Status))
			_, _ = c.Writer.Write(resp.Body)
		} else {
			c.AbortWithStatus(int(resp.Status))
		}
		return false
	}
	return true
}

func buildPluginHandleRequest(c *hertzx.Context) *pluginv1.HandleRequest {
	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	c.Request.ContentLength = int64(len(body))
	hertzx.SyncRequestBody(c)

	query := make(map[string]*pluginv1.StringList)
	if c.Request.URL != nil {
		for k, vs := range c.Request.URL.Query() {
			query[k] = &pluginv1.StringList{Values: vs}
		}
	}

	headers := make(map[string]*pluginv1.StringList)
	for k, vs := range c.Request.Header {
		headers[k] = &pluginv1.StringList{Values: vs}
	}

	reqCtx := &pluginv1.RequestContext{ClientIp: c.ClientIP()}
	if userInfo, ok := c.Request.Context().Value(constant.CtxKeyUser).(*usermodel.User); ok && userInfo != nil {
		reqCtx.UserId = userInfo.Id
		if userInfo.IsAdmin != nil {
			reqCtx.IsAdmin = *userInfo.IsAdmin
		}
	}

	return &pluginv1.HandleRequest{
		Method:  c.Request.Method,
		Path:    c.Request.URL.Path,
		Query:   query,
		Headers: headers,
		Body:    body,
		Context: reqCtx,
	}
}

func writePluginResponse(c *hertzx.Context, resp *pluginv1.HandleResponse) {
	if resp == nil {
		c.Status(http.StatusOK)
		return
	}
	status := int(resp.Status)
	if status == 0 {
		status = http.StatusOK
	}
	for k, v := range resp.Headers {
		c.Header(k, v)
	}
	if len(resp.Body) > 0 {
		c.Writer.WriteHeader(status)
		_, _ = c.Writer.Write(resp.Body)
		return
	}
	c.Status(status)
}
