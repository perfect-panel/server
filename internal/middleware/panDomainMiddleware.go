package middleware

import (
	"net/http"
	"strings"

	"github.com/perfect-panel/server/internal/logic/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
)

func PanDomainMiddleware(svc *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		if svc.Config.Subscribe.PanDomain && c.Request.URL.Path == "/" {
			// intercept browser
			ua := c.GetHeader("User-Agent")

			if svc.Config.Subscribe.UserAgentLimit {
				if !subscribe.IsUserAgentAllowed(c.Request.Context(), svc, ua) {
					c.String(http.StatusForbidden, "Access denied")
					c.Abort()
					return
				}
			}

			domain := c.Request.Host
			domainArr := strings.Split(domain, ".")
			if len(domainArr) < 2 {
				c.String(http.StatusForbidden, "Access denied")
				c.Abort()
				return
			}
			domainFirst := domainArr[0]
			request := types.SubscribeRequest{
				Token: domainFirst,
				Flag:  domainArr[1],
				UA:    c.Request.Header.Get("User-Agent"),
			}
			l := subscribe.NewSubscribeLogic(c.Request.Context(), svc, subscribe.RequestMeta{
				Host:       c.Request.Host,
				RequestURI: c.Request.RequestURI,
				UserAgent:  c.Request.UserAgent(),
				ClientIP:   c.ClientIP(),
			})
			resp, err := l.Handler(&request)
			if err != nil {
				c.String(http.StatusInternalServerError, "Internal Server")
				c.Abort()
				return
			}
			for key, value := range resp.Headers {
				c.Header(key, value)
			}
			c.Header("subscription-userinfo", resp.Header)
			c.String(200, "%s", string(resp.Config))
			c.Abort()
			return
		}
		c.Next()
	}
}
