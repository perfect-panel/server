package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
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
			clientUserAgents := tool.RemoveDuplicateElements(strings.Split(svcCtx.Config.Subscribe.UserAgentList, "\n")...)

			// query client list
			clients, err := svcCtx.ClientModel.List(c.Request.Context())
			if err != nil {
				logger.Errorw("[PanDomainMiddleware] Query client list failed", logger.Field("error", err.Error()))
			}
			for _, item := range clients {
				u := strings.ToLower(item.UserAgent)
				u = strings.Trim(u, " ")
				clientUserAgents = append(clientUserAgents, u)
			}

			var allow = false
			for _, keyword := range clientUserAgents {
				keyword = strings.Trim(keyword, " ")
				if keyword == "" {
					continue
				}
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
			c.String(http.StatusInternalServerError, "Internal Server")
			return
		}
		c.Header("subscription-userinfo", resp.Header)
		c.String(200, "%s", string(resp.Config))
	}
}

func RegisterSubscribeHandlers(router *gin.Engine, serverCtx *svc.ServiceContext) {
	path := serverCtx.Config.Subscribe.SubscribePath
	if path == "" {
		path = "/v1/subscribe/config"
	}
	router.GET(path, SubscribeHandler(serverCtx))
}
