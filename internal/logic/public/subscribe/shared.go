package subscribe

// 跨设备 / 流量加购 / 我的订阅 共享的轻量 helper。
//
// 这些函数/常量原本同 addSubscribeDeviceLogic 放在一起,但拆分 device / trafficaddon
// 模块包后,subscribe.queryMySubscribesLogic 仍需引用 BuildSubscribeURL,而
// device / trafficaddon 仍需引用 RedisServerUserListGlobalKey。把它们留在 subscribe
// 包内,可以让 device / trafficaddon 单向 import subscribe,避免循环依赖。

import (
	"fmt"
	"strings"

	"github.com/perfect-panel/server/internal/svc"
)

// BuildSubscribeURL 拼接 per-device 订阅 URL(决策 14)。
//
// 业务约定(对齐 admin 订阅配置):
//   - SubscribeDomain 多行 = 多域名,取第 1 个非空行(轮询/CDN 由用户自己 DNS 处理)
//   - SubscribePath 配置(如 "/api/subscribe")就用这个,否则兜底 "/v1/subscribe/config"
//   - 域名不含 scheme 时默认 https://
//   - 域名为空 → 返回相对路径,前端按当前 origin 解析(开发兜底)
//
// 多域名场景:同一 token 可生成多条 URL(主线路 / 备用 / CDN 等),
// 前端展示给用户多条选择。这里返回**第一条**作为兼容,需要全部时使用 BuildSubscribeURLs。
func BuildSubscribeURL(svcCtx *svc.ServiceContext, token string) string {
	urls := BuildSubscribeURLs(svcCtx, token)
	if len(urls) == 0 {
		// 域名留空时仍返回一个相对路径,避免前端拿到空字符串。
		path := normalizeSubscribePath(svcCtx.Config.Subscribe.SubscribePath)
		return fmt.Sprintf("%s?token=%s", path, token)
	}
	return urls[0]
}

// BuildSubscribeURLs 把 SubscribeDomain textarea 里**每一行**域名都拼成完整订阅 URL。
//
// 用户在 admin 后台配置多个订阅域名(主线 + 备用 + CDN)时,前端会把所有
// URL 展示给最终用户,让 ta 自己选一条能连通的导入到客户端。返回顺序与
// 配置中的行顺序一致,空行被跳过。
func BuildSubscribeURLs(svcCtx *svc.ServiceContext, token string) []string {
	path := normalizeSubscribePath(svcCtx.Config.Subscribe.SubscribePath)
	domains := splitDomains(svcCtx.Config.Subscribe.SubscribeDomain)
	if len(domains) == 0 {
		return nil
	}
	urls := make([]string, 0, len(domains))
	for _, d := range domains {
		if !strings.HasPrefix(d, "http://") && !strings.HasPrefix(d, "https://") {
			d = "https://" + d
		}
		d = strings.TrimRight(d, "/")
		urls = append(urls, fmt.Sprintf("%s%s?token=%s", d, path, token))
	}
	return urls
}

func normalizeSubscribePath(raw string) string {
	path := strings.TrimSpace(raw)
	if path == "" {
		path = "/v1/subscribe/config"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

// splitDomains — admin 后台「订阅域名」是 textarea(每行一个),返回所有非空行。
func splitDomains(raw string) []string {
	out := make([]string, 0)
	for _, line := range strings.Split(raw, "\n") {
		s := strings.TrimSpace(line)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// pickFirstDomain — 兼容老调用方(暂未使用,保留以防未来需要)。
//
//nolint:unused
func pickFirstDomain(raw string) string {
	d := splitDomains(raw)
	if len(d) == 0 {
		return ""
	}
	return d[0]
}

// RedisServerUserListGlobalKey — 与节点端约定的全局失活 key 前缀(Phase 3 落地)。
// 暂定单 key 模式,所有 server 共用;Phase 3 改为按 server_id 分桶时同步迁移。
var RedisServerUserListGlobalKey = "server:user:list:dirty"
