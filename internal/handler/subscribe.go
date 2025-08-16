package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
)

func SubscribeHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.SubscribeRequest
		if c.Request.Header.Get("token") != "" {
			req.Token = c.Request.Header.Get("token")
		} else {
			req.Token = c.Query("token")
		}
		ua := c.GetHeader("User-Agent")
		req.UA = c.Request.Header.Get("User-Agent")
		req.Flag = c.Query("flag")

		if svcCtx.Config.Subscribe.UserAgentLimit {
			if ua == "" {
				c.String(http.StatusForbidden, "Access denied")
				c.Abort()
				return
			}
			browserKeywords := strings.Split(svcCtx.Config.Subscribe.UserAgentList, "\n")
			var allow = false
			for _, keyword := range browserKeywords {
				if strings.Contains(strings.ToLower(ua), strings.ToLower(keyword)) {
					allow = true
				}
			}
			if !allow {
				c.String(http.StatusForbidden, "Access denied")
				c.Abort()
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
	router.GET(path, SubscribeHandler(serverCtx))
}
