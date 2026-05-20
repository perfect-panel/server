package middleware

import (
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
)

func ServerMiddleware(svc *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		if key, ok := c.GetQuery("secret_key"); ok {
			if key == svc.Config.Node.NodeSecret {
				c.Next()
				return
			}
		}
		c.String(403, "Forbidden")
		c.Abort()
	}
}
