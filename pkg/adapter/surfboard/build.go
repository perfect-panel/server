package surfboard

import (
	"bytes"
	"embed"
	"fmt"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"

	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/traffic"
)

//go:embed *.tpl
var configFiles embed.FS
var shadowsocksSupportMethod = []string{"aes-128-gcm", "aes-192-gcm", "aes-256-gcm", "chacha20-ietf-poly1305"}

func BuildSurfboard(servers proxy.Adapter, siteName string, user UserInfo) []byte {
	var proxies, proxyGroup string
	for _, node := range servers.Proxies {
		if uri := buildProxy(node, user.UUID); uri != "" {
			proxies += uri
		}
	}

	for _, group := range servers.Group {
		if group.Type == proxy.GroupTypeSelect {
			proxyGroup += fmt.Sprintf("%s = select, %s", group.Name, strings.Join(group.Proxies, ", ")) + "\r\n"
		} else if group.Type == proxy.GroupTypeURLTest {
			proxyGroup += fmt.Sprintf("%s = url-test, %s, url=%s, interval=%d", group.Name, strings.Join(group.Proxies, ", "), group.URL, group.Interval) + "\r\n"
		} else if group.Type == proxy.GroupTypeFallback {
			proxyGroup += fmt.Sprintf("%s = fallback, %s, url=%s, interval=%d", group.Name, strings.Join(group.Proxies, ", "), group.URL, group.Interval) + "\r\n"
		} else {
			logger.Errorf("[BuildSurfboard] unknown group type: %s", group.Type)
		}
	}

	var rules string
	for _, rule := range servers.Rules {
		if rule == "" {
			continue
		}
		rules += rule + "\r\n"
	}

	//final rule
	rules += "# 最终规则" + "\r\n" + "FINAL, 手动选择"

	file, err := configFiles.ReadFile("default.tpl")
	if err != nil {
		logger.Errorf("read default surfboard config error: %v", err.Error())
		return nil
	}
	// replace template
	tpl, err := template.New("default").Parse(string(file))
	if err != nil {
		logger.Errorf("read default surfboard config error: %v", err.Error())
		return nil
	}
	var buf bytes.Buffer

	var expiredAt string
	if user.ExpiredDate.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		expiredAt = "长期有效"
	} else {
		expiredAt = user.ExpiredDate.Format("2006-01-02 15:04:05")
	}
	// convert traffic
	upload := traffic.AutoConvert(user.Upload, false)
	download := traffic.AutoConvert(user.Download, false)
	total := traffic.AutoConvert(user.TotalTraffic, false)
	unusedTraffic := traffic.AutoConvert(user.TotalTraffic-user.Upload-user.Download, false)
	// query Host
	urlParse, err := url.Parse(user.SubscribeURL)
	if err != nil {
		return nil
	}
	if err := tpl.Execute(&buf, map[string]interface{}{
		"Proxies":         proxies,
		"ProxyGroup":      proxyGroup,
		"SubscribeURL":    user.SubscribeURL,
		"SubscribeInfo":   fmt.Sprintf("title=%s订阅信息, content=上传流量：%s\\n下载流量：%s\\n剩余流量: %s\\n套餐流量：%s\\n到期时间：%s", siteName, upload, download, unusedTraffic, total, expiredAt),
		"SubscribeDomain": urlParse.Host,
		"Rules":           rules,
	}); err != nil {
		logger.Errorf("build surfboard config error: %v", err.Error())
		return nil
	}
	return buf.Bytes()
}

func buildProxy(data proxy.Proxy, uuid string) string {
	var p string
	switch data.Protocol {
	case "vmess":
		p = buildVMess(data, uuid)
	case "shadowsocks":
		if !tool.Contains(shadowsocksSupportMethod, data.Option.(proxy.Shadowsocks).Method) {
			return ""
		}
		p = buildShadowsocks(data, uuid)
	case "trojan":
		p = buildTrojan(data, uuid)
	}
	return p
}
