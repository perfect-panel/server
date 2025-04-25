package surfboard

import (
	"fmt"
	"strings"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

func buildVMess(data proxy.Proxy, uuid string) string {
	vmess, ok := data.Option.(proxy.Vmess)
	if !ok {
		return ""
	}
	addr := fmt.Sprintf("%s=vmess, %s, %d", data.Name, data.Server, data.Port)
	uriConfig := []string{
		addr,
		fmt.Sprintf("username=%s", uuid),
		"vmess-aead=true",
		"tfo=true",
		"udp-relay=true",
	}
	if vmess.Security == "tls" {
		uriConfig = append(uriConfig, "tls=true")
		if vmess.SecurityConfig.AllowInsecure {
			uriConfig = append(uriConfig, "skip-cert-verify=true")
		} else {
			uriConfig = append(uriConfig, "skip-cert-verify=false")
		}
		if vmess.SecurityConfig.SNI != "" {
			uriConfig = append(uriConfig, fmt.Sprintf("sni=%s", vmess.SecurityConfig.SNI))
		}
	}
	if vmess.Transport == "websocket" {
		uriConfig = append(uriConfig, "ws=true")
		if vmess.TransportConfig.Path != "" {
			uriConfig = append(uriConfig, fmt.Sprintf("ws-path=%s", vmess.TransportConfig.Path))
		}
		if vmess.TransportConfig.Host != "" {
			uriConfig = append(uriConfig, fmt.Sprintf("ws-headers=Host:%s", vmess.TransportConfig.Host))
		}
	}

	return strings.Join(uriConfig, ",") + "\r\n"
}
