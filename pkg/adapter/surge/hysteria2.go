package surge

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

func buildHysteria2(data proxy.Proxy, uuid string) string {
	hysteria2, ok := data.Option.(proxy.Hysteria2)
	if !ok {
		return ""
	}

	var port int
	if hysteria2.HopPorts != "" {
		ports := strings.Split(hysteria2.HopPorts, ",")
		p := ports[0]
		if len(strings.Split(p, "-")) > 1 {
			p = strings.Split(p, "-")[0]
		}
		port, _ = strconv.Atoi(p)
	} else {
		port = data.Port
	}

	config := []string{
		fmt.Sprintf("%s=hysteria2,%s,%d", data.Name, data.Server, port),
		"password=" + uuid,
		"udp-relay=true",
	}
	if hysteria2.SecurityConfig.SNI != "" {
		config = append(config, "sni="+hysteria2.SecurityConfig.SNI)
	}
	if hysteria2.SecurityConfig.AllowInsecure {
		config = append(config, "skip-cert-verify=true")
	} else {
		config = append(config, "skip-cert-verify=false")
	}
	return strings.Join(config, ",") + "\r\n"
}
