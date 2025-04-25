package surfboard

import (
	"testing"

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

func TestVMess(t *testing.T) {
	node := createVMess()
	uuid := "123456"
	p := buildVMess(node, uuid)
	t.Log(p)
}
