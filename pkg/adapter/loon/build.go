package loon

import (
	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

func BuildLoon(servers []proxy.Proxy, uuid string) []byte {
	uri := ""
	for _, s := range servers {
		switch s.Protocol {
		case "vmess":
			uri += buildVMess(s, uuid)
		case "shadowsocks":
			uri += buildShadowsocks(s, uuid)
		case "trojan":
			uri += buildTrojan(s, uuid)
		case "vless":
			uri += buildVless(s, uuid)
		case "hysteria2":
			uri += buildHysteria2(s, uuid)
		default:
			continue
		}
	}

	return []byte(uri)
}
