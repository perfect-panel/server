package surfboard

import (
	"testing"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

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

func TestTrojan(t *testing.T) {
	node := createTrojan()
	uuid := "123456"
	trojan := buildTrojan(node, uuid)
	t.Log(trojan)
}
