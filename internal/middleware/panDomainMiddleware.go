package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/tool"
)

func PanDomainMiddleware(svc *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		if svc.Config.Subscribe.PanDomain && c.Request.URL.Path == "/" {
			// intercept browser
			ua := c.GetHeader("User-Agent")

			if svc.Config.Subscribe.UserAgentLimit {
				if ua == "" {
					c.String(http.StatusForbidden, "Access denied")
					c.Abort()
					return
				}
				browserKeywords := tool.RemoveDuplicateElements(strings.Split(svc.Config.Subscribe.UserAgentList, "\n")...)
				var allow = false
				for _, keyword := range browserKeywords {
					keyword = strings.ToLower(strings.Trim(keyword, " "))
					if keyword == "" {
						continue
					}
					if strings.Contains(strings.ToLower(ua), keyword) {
						allow = true
					}
				}
				if !allow {
					c.String(http.StatusForbidden, "Access denied")
					c.Abort()
					return
				}
			}

			domain := c.Request.Host
			domainArr := strings.Split(domain, ".")
			domainFirst := domainArr[0]
			request := types.SubscribeRequest{
				Token: domainFirst,
				Flag:  domainArr[1],
				UA:    c.Request.Header.Get("User-Agent"),
			}
			l := subscribe.NewSubscribeLogic(c, svc)
			resp, err := l.Handler(&request)
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
