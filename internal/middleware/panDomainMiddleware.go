package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
)

func PanDomainMiddleware(svc *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		domain := c.Request.Host
		domainArr := strings.Split(domain, ".")
		domainFirst := domainArr[0]
		request := types.SubscribeRequest{
			Token: domainFirst,
			Flag:  domainArr[1],
			UA:    c.Request.Header.Get("User-Agent"),
		}

		if svc.Config.Subscribe.PanDomain && len(domainFirst) == 32 {
			// intercept browser
			ua := c.GetHeader("User-Agent")
			if ua == "" {
				c.String(http.StatusForbidden, "Access denied")
				c.Abort()
				return
			}
			browserKeywords := []string{"chrome", "firefox", "safari", "edge", "opera", "micromessenger"}
			for _, keyword := range browserKeywords {
				lcUA := strings.ToLower(ua)
				if strings.Contains(lcUA, keyword) {
					c.String(http.StatusForbidden, "Access denied")
					c.Abort()
					return
				}
			}

			l := subscribe.NewSubscribeLogic(c, svc)
			resp, err := l.Generate(&request)
			if err != nil {
				return
			}
			c.Header("subscription-userinfo", resp.Header)
			c.String(200, "%s", string(resp.Config))
			c.Abort()
			return
		}
		c.Next()
	}
}
