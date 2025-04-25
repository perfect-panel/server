package quantumultx

import (
	"fmt"
	"strings"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

func buildVmess(data proxy.Proxy, uuid string) string {

	vmess := data.Option.(proxy.Vmess)
	addr := fmt.Sprintf("vmess=%s:%d", data.Server, data.Port)
	var host string
	uriConfig := []string{
		addr,
		"method=chacha20-poly1305",
		fmt.Sprintf("password=%s", uuid),
		"fast-open=true",
		"udp-relay=true",
		fmt.Sprintf("tag=%s", data.Name),
	}
	if vmess.Security == "tls" {
		if vmess.Transport == "tcp" {
			uriConfig = append(uriConfig, "obfs=over-tls")
		}
		if vmess.SecurityConfig.AllowInsecure {
			uriConfig = append(uriConfig, "tls-verification=true")
		} else {
			uriConfig = append(uriConfig, "tls-verification=false")
		}
		if vmess.SecurityConfig.SNI != "" {
			host = vmess.SecurityConfig.SNI
		}
	}

	if vmess.Transport == "websocket" {
		uriConfig = append(uriConfig, fmt.Sprintf("obfs-uri=%s", vmess.TransportConfig.Path))
		host = vmess.TransportConfig.Host
	}
	if host != "" {
		uriConfig = append(uriConfig, fmt.Sprintf("obfs-host=%s", host))
	}
	return strings.Join(uriConfig, ",") + "\r\n"
}
