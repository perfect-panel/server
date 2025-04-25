package loon

import (
	"testing"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

func createSS() proxy.Proxy {
	return proxy.Proxy{
		Name:     "Shadowsocks",
		Server:   "127.0.0.1",
		Port:     10301,
		Protocol: "shadowsocks",
		Option: proxy.Shadowsocks{
			Method:    "aes-256-gcm",
			ServerKey: "",
		},
	}

}

func TestBuildSS(t *testing.T) {
	s := createSS()

	password := "f0d0237d-193a-4cf5-99dd-b02207beaea6"
	uri := buildShadowsocks(s, password)
	t.Log(uri)
}
