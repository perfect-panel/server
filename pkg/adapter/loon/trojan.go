package loon

import (
	"fmt"
	"strings"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

func buildTrojan(data proxy.Proxy, password string) string {
	trojan := data.Option.(proxy.Trojan)

	configs := []string{
		fmt.Sprintf("%s=trojan", data.Name),
		data.Server,
		fmt.Sprintf("%d", data.Port),
		"auto",
		password,
		"fast-open=false",
		"udp=true",
	}

	if trojan.SecurityConfig.SNI != "" {
		configs = append(configs, fmt.Sprintf("sni=%s", trojan.SecurityConfig.SNI))
	}
	if trojan.SecurityConfig.AllowInsecure {
		configs = append(configs, "skip-cert-verify=true")
	} else {
		configs = append(configs, "skip-cert-verify=false")
	}

	if trojan.Transport == "websocket" {
		configs = append(configs, "transport=ws")
		if trojan.TransportConfig.Path != "" {
			configs = append(configs, fmt.Sprintf("path=%s", trojan.TransportConfig.Path))
		}
		if trojan.TransportConfig.Host != "" {
			configs = append(configs, fmt.Sprintf("host=%s", trojan.TransportConfig.Host))
		}
	}

	uri := strings.Join(configs, ",")
	return uri + "\r\n"
}
