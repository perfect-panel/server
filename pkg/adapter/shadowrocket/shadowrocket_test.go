package shadowrocket

import (
	"testing"
	"time"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

func createVMess() proxy.Proxy {
	return proxy.Proxy{
		Name:     "Vmess",
		Server:   "test.xxxx.com",
		Port:     13002,
		Protocol: "vmess",
		Option: proxy.Vmess{
			Port:      13002,
			Transport: "websocket",
			TransportConfig: proxy.TransportConfig{
				Path: "/ws",
				Host: "test.xx.com",
			},
			Security: "none",
		},
	}
}

func createSS() proxy.Proxy {
	return proxy.Proxy{
		Name:     "Shadowsocks",
		Server:   "test.xxxx.com",
		Port:     10301,
		Protocol: "shadowsocks",
		Option: proxy.Shadowsocks{
			Port:      10301,
			Method:    "aes-256-gcm",
			ServerKey: "123456",
		},
	}
}

func createTrojan() proxy.Proxy {

	return proxy.Proxy{
		Name:     "Trojan",
		Server:   "test.xxxx.com",
		Port:     13002,
		Protocol: "trojan",
		Option: proxy.Trojan{
			Port:      13002,
			Transport: "websocket",
			TransportConfig: proxy.TransportConfig{
				Path: "/ws",
				Host: "baidu.com",
			},
			SecurityConfig: proxy.SecurityConfig{
				SNI:           "baidu.com",
				AllowInsecure: true,
			},
		},
	}
}
func TestBuildShadowrocket(t *testing.T) {
	s := []proxy.Proxy{
		createVMess(),
		createSS(),
		createTrojan(),
	}
	uri := BuildShadowrocket(s, "uuid", UserInfo{
		Upload:       1024,
		Download:     1024,
		TotalTraffic: 2048,
		ExpiredDate:  time.Now().AddDate(0, 0, 1),
	})
	t.Log(string(uri))
}
