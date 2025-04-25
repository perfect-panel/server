package quantumultx

import (
	"fmt"
	"strings"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

// 生成 Trojan 配置
func buildTrojan(data proxy.Proxy, password string) string {
	trojan := data.Option.(proxy.Trojan)

	addr := fmt.Sprintf("trojan=%s:%d", data.Server, data.Port)
	config := []string{
		addr,
		fmt.Sprintf("password=%s", password),
		"fast-open=true",
		"udp-relay=true",
		fmt.Sprintf("tag=%s", data.Name),
	}

	if trojan.Transport == "websocket" {
		config = append(config, "obfs=wss")
		if trojan.TransportConfig.Path != "" {
			config = append(config, fmt.Sprintf("obfs-uri=%s", trojan.TransportConfig.Path))
		}
		if trojan.TransportConfig.Host != "" {
			config = append(config, fmt.Sprintf("obfs-host=%s", trojan.TransportConfig.Host))
		}
	} else {
		config = append(config, "over-tls=true")
		if trojan.SecurityConfig.SNI != "" {
			config = append(config, fmt.Sprintf("tls-host=%s", trojan.SecurityConfig.SNI))
		}
	}

	return strings.Join(config, ",") + "\r\n"
}
