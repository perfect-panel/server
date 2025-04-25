package surfboard

import (
	"testing"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

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

func TestShadowsocks(t *testing.T) {
	node := createSS()
	uuid := "123456"
	shadowsocks := buildShadowsocks(node, uuid)
	t.Log(shadowsocks)
}
