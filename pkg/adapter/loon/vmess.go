package loon

import (
	"fmt"
	"strings"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
)

func buildVMess(data proxy.Proxy, password string) string {
	vmess := data.Option.(proxy.Vmess)

	configs := []string{
		fmt.Sprintf("%s=vmess", data.Name),
		data.Server,
		fmt.Sprintf("%d", data.Port),
		"auto",
		password,
		"fast-open=false",
		"udp=true",
		"alterId=0",
	}

	switch vmess.Transport {
	case "tcp":
		configs = append(configs, "transport=tcp")
	case "websocket":
		configs = append(configs, "transport=ws")
		if vmess.TransportConfig.Path != "" {
			configs = append(configs, fmt.Sprintf("path=%s", vmess.TransportConfig.Path))
		}
		if vmess.TransportConfig.Host != "" {
			configs = append(configs, fmt.Sprintf("host=%s", vmess.TransportConfig.Host))
		}
	default:
		logger.Info("Loon Unknown transport type: ", logger.Field("transport", vmess.Transport))
		return ""
	}

	if vmess.Security == "tls" {
		configs = append(configs, "over-tls=true", fmt.Sprintf("tls-name=%s", vmess.SecurityConfig.SNI))
		if vmess.SecurityConfig.AllowInsecure {
			configs = append(configs, "skip-cert-verify=true")
		} else {
			configs = append(configs, "skip-cert-verify=false")
		}

	}

	uri := strings.Join(configs, ",")
	return uri + "\r\n"
}
