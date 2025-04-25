package loon

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"

	"github.com/perfect-panel/ppanel-server/pkg/logger"
)

func buildVless(data proxy.Proxy, password string) string {
	vless := data.Option.(proxy.Vless)
	// If flow is not empty, it means that the server is a relay server
	if vless.Flow != "" {
		return ""
	}

	configs := []string{
		fmt.Sprintf("%s=vless", data.Name),
		data.Server,
		strconv.Itoa(data.Port),
		"auto",
		password,
		"fast-open=false",
		"udp=true",
		"alterId=0",
	}

	switch vless.Transport {
	case "tcp":
		configs = append(configs, "transport=tcp")
	case "websocket":
		configs = append(configs, "transport=ws")
		if vless.TransportConfig.Path != "" {
			configs = append(configs, fmt.Sprintf("path=%s", vless.TransportConfig.Path))
		}
		if vless.TransportConfig.Host != "" {
			configs = append(configs, fmt.Sprintf("host=%s", vless.TransportConfig.Host))
		}
	default:
		logger.Info("Loon Unknown transport type: ", logger.Field("transport", vless.Transport))
		return ""
	}

	if vless.Security == "tls" {
		configs = append(configs, "over-tls=true", fmt.Sprintf("tls-name=%s", vless.SecurityConfig.SNI))
		if vless.SecurityConfig.AllowInsecure {
			configs = append(configs, "skip-cert-verify=true")
		} else {
			configs = append(configs, "skip-cert-verify=false")
		}
	} else if vless.Security == "reality" {
		// Loon does not support reality security
		logger.Info("Loon Unknown security type: ", logger.Field("security", vless.Security))
		return ""
	}

	uri := strings.Join(configs, ",")
	return uri + "\r\n"
}
