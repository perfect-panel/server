package singbox

import (
	"testing"

	"github.com/perfect-panel/server/pkg/adapter/proxy"

	"github.com/stretchr/testify/assert"
)

func createSS() proxy.Proxy {
	c := proxy.Shadowsocks{
		Method:    "aes-256-gcm",
		Port:      10301,
		ServerKey: "",
	}
	return proxy.Proxy{
		Name:     "Shadowsocks",
		Server:   "127.0.0.1",
		Port:     10301,
		Protocol: "shadowsocks",
		Option:   c,
	}
}

func createVLESS() proxy.Proxy {
	c := proxy.Vless{
		Port:      10301,
		Flow:      "xtls-rprx-direct",
		Transport: "websocket",
		TransportConfig: proxy.TransportConfig{
			Path: "/ws",
			Host: "baidu.com",
		},
		Security: "tls",
		SecurityConfig: proxy.SecurityConfig{
			SNI:           "baidu.com",
			Fingerprint:   "chrome",
			AllowInsecure: true,
		},
	}
	s := proxy.Proxy{
		Name:     "VLESS",
		Server:   "test.xxx.com",
		Port:     10301,
		Protocol: "vless",
		Option:   c,
	}
	return s
}

func TestSingboxShadowsocks(t *testing.T) {
	s := createSS()
	p, err := ParseShadowsocks(s, "uuid")
	if err != nil {
		t.Fatal(err)
	}
	data, err := p.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, 0, len(data))

	// Output:
	// proxy: proxy: {"tag":"Shadowsocks","type":"shadowsocks","server":"127.0.0.1","server_port":10301,"method":"aes-256-gcm","password":"uuid","network":"tcp"}

}

func TestSingboxVless(t *testing.T) {
	s := createVLESS()
	p, err := ParseVless(s, "uuid")
	if err != nil {
		t.Fatal(err)
	}
	data, err := p.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, 0, len(data))
}
