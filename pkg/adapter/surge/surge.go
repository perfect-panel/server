package surge

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
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

func (m *Surge) Build(siteName string, user UserInfo) []byte {
	var proxies, proxyGroup string
	var removed []string
	var ps []string

	for _, p := range m.Adapter.Proxies {
		switch p.Protocol {
		case "shadowsocks":
			proxies += buildShadowsocks(p, user.UUID)
		case "trojan":
			proxies += buildTrojan(p, user.UUID)
		case "hysteria2":
			proxies += buildHysteria2(p, user.UUID)
		case "vmess":
			proxies += buildVMess(p, user.UUID)
		default:
			removed = append(removed, p.Name)
		}
		ps = append(ps, p.Name)
	}

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

	ps = tool.RemoveStringElement(ps, removed...)
	proxyGroup = strings.Join(ps, ",")

	// convert traffic
	upload := traffic.AutoConvert(user.Upload, false)
	download := traffic.AutoConvert(user.Download, false)
	total := traffic.AutoConvert(user.TotalTraffic, false)
	unusedTraffic := traffic.AutoConvert(user.TotalTraffic-user.Upload-user.Download, false)
	// query Host
	if err := tpl.Execute(&buf, map[string]interface{}{
		"Proxies":       proxies,
		"ProxyGroup":    proxyGroup,
		"SubscribeURL":  user.SubscribeURL,
		"SubscribeInfo": fmt.Sprintf("title=%s订阅信息, content=上传流量：%s\\n下载流量：%s\\n剩余流量: %s\\n套餐流量：%s\\n到期时间：%s", siteName, upload, download, unusedTraffic, total, expiredAt),
	}); err != nil {
		logger.Errorf("build Surge config error: %v", err.Error())
		return nil
	}
	return buf.Bytes()
}
