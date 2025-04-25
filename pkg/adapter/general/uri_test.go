package general

import (
	"testing"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

func createServer() proxy.Proxy {
	return proxy.Proxy{
		Name:     "Meta",
		Server:   "127.0.0.1",
		Port:     13092,
		Protocol: "shadowsocks",
		Option: proxy.Shadowsocks{
			Method:    "aes-256-gcm",
			ServerKey: "",
		},
	}
}

func TestGenerateBase64General(t *testing.T) {
	s := createServer()
	p := buildProxy(s, "935b33c7-e128-49f2-816b-71070469cac2")
	t.Log(p)
}
