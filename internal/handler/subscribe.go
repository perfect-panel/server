package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
)

func V2SubscribeHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.SubscribeRequest
		if c.Request.Header.Get("token") != "" {
			req.Token = c.Request.Header.Get("token")
		} else {
			req.Token = c.Query("token")
		}
		req.UA = c.Request.Header.Get("User-Agent")
		req.Flag = c.Query("flag")

		// intercept browser
		ua := c.GetHeader("User-Agent")
		if ua == "" {
			c.String(http.StatusForbidden, "Access denied")
			return
		}
		browserKeywords := []string{"chrome", "firefox", "safari", "edge", "opera", "micromessenger"}
		for _, keyword := range browserKeywords {
			lcUA := strings.ToLower(ua)
			if strings.Contains(lcUA, keyword) {
				c.String(http.StatusForbidden, "Access denied")
				return
			}
		}

		l := subscribe.NewSubscribeLogic(c, svcCtx)
		resp, err := l.Handler(&req)
		if err != nil {
			return
		}
		c.Header("subscription-userinfo", resp.Header)
		c.String(200, "%s", string(resp.Config))
	}
}

func RegisterSubscribeHandlers(router *gin.Engine, serverCtx *svc.ServiceContext) {
	path := serverCtx.Config.Subscribe.SubscribePath
	if path == "" {
		path = "/api/subscribe"
	}
	router.GET(path, V2SubscribeHandler(serverCtx))

	router.GET(path+"/v2", V2SubscribeHandler(serverCtx))
}
