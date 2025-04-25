package quantumultx

import (
	"encoding/base64"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

func BuildQuantumultX(servers []proxy.Proxy, uuid string) string {
	var uri string
	for _, s := range servers {
		switch s.Protocol {
		case "vmess":
			uri += buildVmess(s, uuid)
		case "shadowsocks":
			uri += buildShadowsocks(s, uuid)
		case "trojan":
			uri += buildTrojan(s, uuid)
		}
	}
	return base64.StdEncoding.EncodeToString([]byte(uri))
}
