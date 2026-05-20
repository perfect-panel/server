package handler

import (
	"net/http"
	"strings"

	"github.com/perfect-panel/server/internal/logic/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
)

func SubscribeHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.SubscribeRequest
		if c.Request.Header.Get("token") != "" {
			req.Token = c.Request.Header.Get("token")
		} else {
			req.Token = c.Query("token")
		}
		ua := c.GetHeader("User-Agent")
		req.UA = c.Request.Header.Get("User-Agent")
		req.Flag = c.Query("flag")
		req.Type = c.Query("type")
		// 获取所有查询参数
		req.Params = getQueryMap(c.Request)

		if svcCtx.Config.Subscribe.PanDomain {
			domain := c.Request.Host
			domainArr := strings.Split(domain, ".")
			short, err := tool.FixedUniqueString(req.Token, 8, "")
			if err != nil {
				logger.Errorf("[SubscribeHandler] Generate short token failed: %v", err)
				c.String(http.StatusInternalServerError, "Internal Server")
				c.Abort()
				return
			}
			if strings.ToLower(short) != strings.ToLower(domainArr[0]) {
				logger.Debugf("[SubscribeHandler] Generate short token failed, short: %s, domain: %s", short, domainArr[0])
				c.String(http.StatusForbidden, "Access denied")
				c.Abort()
				return
			}
		}

		if svcCtx.Config.Subscribe.UserAgentLimit {
			if !subscribe.IsUserAgentAllowed(c.Request.Context(), svcCtx, ua) {
				c.String(http.StatusForbidden, "Access denied")
				c.Abort()
				return
			}
		}

		l := subscribe.NewSubscribeLogic(c.Request.Context(), svcCtx, subscribe.RequestMeta{
			Host:       c.Request.Host,
			RequestURI: c.Request.RequestURI,
			UserAgent:  c.Request.UserAgent(),
			ClientIP:   c.ClientIP(),
		})
		resp, err := l.Handler(&req)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server")
			return
		}
		for key, value := range resp.Headers {
			c.Header(key, value)
		}
		c.Header("subscription-userinfo", resp.Header)
		c.String(200, "%s", string(resp.Config))
	}
}

func RegisterSubscribeHandlers(router *hertzx.Engine, serverCtx *svc.ServiceContext) {
	path := serverCtx.Config.Subscribe.SubscribePath
	if path == "" {
		path = "/v1/subscribe/config"
	}
	router.GET(path, SubscribeHandler(serverCtx))
	if serverCtx.Config.Subscribe.PanDomain {
		router.GET("/", PanDomainSubscribeHandler(serverCtx))
	}
}

func PanDomainSubscribeHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		ua := c.GetHeader("User-Agent")
		if svcCtx.Config.Subscribe.UserAgentLimit && !subscribe.IsUserAgentAllowed(c.Request.Context(), svcCtx, ua) {
			c.String(http.StatusForbidden, "Access denied")
			c.Abort()
			return
		}

		domainArr := strings.Split(c.Request.Host, ".")
		if len(domainArr) < 2 {
			c.String(http.StatusForbidden, "Access denied")
			c.Abort()
			return
		}

		req := types.SubscribeRequest{
			Token:  domainArr[0],
			Flag:   domainArr[1],
			UA:     ua,
			Params: getQueryMap(c.Request),
		}
		l := subscribe.NewSubscribeLogic(c.Request.Context(), svcCtx, subscribe.RequestMeta{
			Host:       c.Request.Host,
			RequestURI: c.Request.RequestURI,
			UserAgent:  c.Request.UserAgent(),
			ClientIP:   c.ClientIP(),
		})
		resp, err := l.Handler(&req)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server")
			return
		}
		for key, value := range resp.Headers {
			c.Header(key, value)
		}
		c.Header("subscription-userinfo", resp.Header)
		c.String(http.StatusOK, "%s", string(resp.Config))
	}
}

// GetQueryMap 将 http.Request 的查询参数转换为 map[string]string
func getQueryMap(r *http.Request) map[string]string {
	result := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
