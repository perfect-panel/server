package surge

import (
	"bytes"
	"embed"
	"fmt"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/traffic"
)

//go:embed *.tpl
var configFiles embed.FS

type UserInfo struct {
	UUID         string
	Upload       int64
	Download     int64
	TotalTraffic int64
	ExpiredDate  time.Time
	SubscribeURL string
}

type Surge struct {
	Adapter proxy.Adapter
	UUID    string
	User    UserInfo
}

func NewSurge(adapter proxy.Adapter) *Surge {
	return &Surge{
		Adapter: adapter,
	}
}

func (m *Surge) Build(uuid, siteName string, user UserInfo) []byte {
	var proxies, proxyGroup, rules string

	for _, p := range m.Adapter.Proxies {
		switch p.Protocol {
		case "shadowsocks":
			proxies += buildShadowsocks(p, uuid)
		case "trojan":
			proxies += buildTrojan(p, uuid)
		case "hysteria2":
			proxies += buildHysteria2(p, uuid)
		case "vmess":
			proxies += buildVMess(p, uuid)
		}
	}
	for _, group := range m.Adapter.Group {
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
	for _, rule := range m.Adapter.Rules {
		if rule == "" {
			continue
		}
		rules += rule + "\r\n"
	}
	//final rule
	rules += "# 最终规则" + "\r\n" + "FINAL,手动选择,dns-failed"

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
		logger.Errorf("build Surge config error: %v", err.Error())
		return nil
	}
	return buf.Bytes()
}
