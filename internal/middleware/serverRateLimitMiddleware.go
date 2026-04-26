package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
)

// ServerRateLimitWindow is the rolling window length for per-(server_id, endpoint)
// call counting. 1 second keeps bookkeeping cheap while giving a stable QPS signal.
const ServerRateLimitWindow = time.Second

// ServerRateLimitMaxPerWindow caps the number of accepted requests per
// (server_id, endpoint) within one window. Set generously so legitimate push
// traffic is not throttled; intended as a hard ceiling against runaway nodes.
const ServerRateLimitMaxPerWindow = 20

// serverRateLimitKey prefixes a Redis counter scoped by node and endpoint.
const serverRateLimitKey = "node:ratelimit:%s:%s"

// ServerRateLimitMiddleware rejects a node that exceeds ServerRateLimitMaxPerWindow
// requests per second on the same endpoint. Keyed by server_id (from query)
// plus the route template, so heavy /online traffic does not starve /reject.
//
// Fail-open: if Redis is unreachable we let the request through rather than
// dropping push/report traffic — degraded observability > denial-of-service.
func ServerRateLimitMiddleware(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		serverID := c.Query("server_id")
		if serverID == "" {
			c.Next()
			return
		}
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		key := fmt.Sprintf(serverRateLimitKey, serverID, endpoint)
		ctx := c.Request.Context()

		count, err := svcCtx.Redis.Incr(ctx, key).Result()
		if err != nil {
			logger.Errorw("[ServerRateLimit] redis incr failed, fail-open",
				logger.Field("key", key),
				logger.Field("error", err.Error()))
			c.Next()
			return
		}
		if count == 1 {
			// First hit in the window; arm expiry so the key self-cleans.
			_ = svcCtx.Redis.Expire(ctx, key, ServerRateLimitWindow).Err()
		}
		if count > ServerRateLimitMaxPerWindow {
			logger.Errorw("[ServerRateLimit] rate limit exceeded",
				logger.Field("server_id", serverID),
				logger.Field("endpoint", endpoint),
				logger.Field("count", count))
			c.String(429, "Too Many Requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
