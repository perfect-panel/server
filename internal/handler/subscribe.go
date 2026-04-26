package handler

import (
	"fmt"
	"net/http"
	"strconv"
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

		// V4.3:智能下发"自动更新间隔"。
		//
		//   - Profile-Update-Interval(单位:小时):Clash 家族 / Hiddify / Mihomo Party 等读
		//   - #!MANAGED-CONFIG <url> interval=<秒> strict=false:Surge / Stash 必须放配置体首行
		//
		// 0 = admin 关闭该功能,header 不下发,Surge 也不注入。
		intervalHours := svcCtx.Config.Subscribe.UpdateIntervalHours
		body := resp.Config
		if intervalHours > 0 {
			c.Header("Profile-Update-Interval", strconv.FormatInt(intervalHours, 10))
			if isSurgeFamily(ua) {
				body = injectSurgeManagedConfig(body, c.Request, intervalHours)
			}
		}

		c.String(200, "%s", string(body))
	}
}

// isSurgeFamily — Surge / Surge for Mac / Stash 都识别 #!MANAGED-CONFIG。
// 通过 UA 子串匹配,大小写不敏感。其他客户端走 header 即可。
func isSurgeFamily(ua string) bool {
	u := strings.ToLower(ua)
	return strings.Contains(u, "surge") || strings.Contains(u, "stash")
}

// injectSurgeManagedConfig — 在 Surge 配置文件首行插入 #!MANAGED-CONFIG。
//
// 规范(Surge 文档):
//   #!MANAGED-CONFIG <url> interval=<seconds> strict=<true|false>
//   - url:订阅 URL,Surge 按这个 URL 拉取最新配置
//   - interval:轮询周期,单位**秒**(注意不是小时)
//   - strict:true=拉取失败保持旧配置;我们用 false,允许失败时回退本地
//
// 已经存在 #!MANAGED-CONFIG 时(用户的模板自己写了)就跳过,不重复注入。
func injectSurgeManagedConfig(body []byte, r *http.Request, hours int64) []byte {
	if len(body) == 0 {
		return body
	}
	// 已存在则尊重原配置,不覆盖
	head := body
	if len(head) > 256 {
		head = head[:256]
	}
	if strings.Contains(strings.ToLower(string(head)), "#!managed-config") {
		return body
	}
	// 还原完整订阅 URL(兼容反代)
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") == "" {
		scheme = "http"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	host := r.Host
	if h := r.Header.Get("X-Forwarded-Host"); h != "" {
		host = h
	}
	url := fmt.Sprintf("%s://%s%s", scheme, host, r.URL.RequestURI())
	directive := fmt.Sprintf("#!MANAGED-CONFIG %s interval=%d strict=false\n", url, hours*3600)
	out := make([]byte, 0, len(directive)+len(body))
	out = append(out, []byte(directive)...)
	out = append(out, body...)
	return out
}

func RegisterSubscribeHandlers(router *gin.Engine, serverCtx *svc.ServiceContext) {
	path := serverCtx.Config.Subscribe.SubscribePath
	if path == "" {
		path = "/v1/subscribe/config"
	}
	router.GET(path, SubscribeHandler(serverCtx))
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
