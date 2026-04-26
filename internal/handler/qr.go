package handler

// V4.3 决策 14:每订阅 URL 自动生成二维码。
// 端点:GET /v1/public/qr?token=<token>&size=256
// - 校验 token 存在(device 或 legacy user_subscribe)
// - 拼接绝对订阅 URL = <SubscribeDomain or req.Host><SubscribePath>?token=<token>
// - 返回 image/png

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

const (
	qrDefaultSize = 256
	qrMinSize     = 64
	qrMaxSize     = 1024
)

func QRHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Query("token"))
		if token == "" {
			c.String(http.StatusBadRequest, "missing token")
			return
		}

		// 校验 token:V4.3 优先匹 device.token,fallback 到 legacy user_subscribe.token。
		if !verifyToken(c, svcCtx, token) {
			c.String(http.StatusNotFound, "token not found")
			return
		}

		size := parseSize(c.Query("size"))
		fullURL := buildSubscribeURL(c, svcCtx, token)

		png, err := qrcode.Encode(fullURL, qrcode.Medium, size)
		if err != nil {
			logger.Errorf("[QRHandler] encode failed: %v", err)
			c.String(http.StatusInternalServerError, "encode failed")
			return
		}
		// 缓存 1 小时(token 重置后会换新 token,旧 URL 自然失效)
		c.Header("Cache-Control", "public, max-age=3600")
		c.Header("Content-Type", "image/png")
		c.Writer.Write(png)
	}
}

func verifyToken(c *gin.Context, svcCtx *svc.ServiceContext, token string) bool {
	ctx := c.Request.Context()
	if _, err := svcCtx.UserModel.FindOneSubscribeDeviceByToken(ctx, token); err == nil {
		return true
	}
	// legacy fallback:旧机场或未迁移用户
	_, err := svcCtx.UserModel.FindOneSubscribeByToken(ctx, token)
	if err == nil {
		return true
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf("[QRHandler] token lookup failed: %v", err)
	}
	return false
}

// buildSubscribeURL 拼接客户端订阅 URL,优先用配置的 SubscribeDomain;
// 否则按当前请求 Host + 反代 Proto 推断。
func buildSubscribeURL(c *gin.Context, svcCtx *svc.ServiceContext, token string) string {
	path := svcCtx.Config.Subscribe.SubscribePath
	if path == "" {
		path = "/v1/subscribe/config"
	}
	domain := strings.TrimSpace(svcCtx.Config.Subscribe.SubscribeDomain)
	var base string
	if domain != "" {
		// SubscribeDomain 可能含/不含 scheme,统一规整
		if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
			base = strings.TrimRight(domain, "/")
		} else {
			base = "https://" + strings.TrimRight(domain, "/")
		}
	} else {
		scheme := "https"
		if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		} else if c.Request.TLS == nil && c.Request.Host != "" && !strings.Contains(c.Request.Host, ":443") {
			scheme = "http"
		}
		base = scheme + "://" + c.Request.Host
	}
	return base + path + "?token=" + token
}

func parseSize(raw string) int {
	if raw == "" {
		return qrDefaultSize
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return qrDefaultSize
	}
	if n < qrMinSize {
		return qrMinSize
	}
	if n > qrMaxSize {
		return qrMaxSize
	}
	return n
}
