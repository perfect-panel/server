package loon

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

func buildHysteria2(data proxy.Proxy, password string) string {
	hysteria2 := data.Option.(proxy.Hysteria2)

	configs := []string{
		fmt.Sprintf("%s=Hysteria2", data.Name),
		data.Server,
		strconv.Itoa(data.Port),
		password,
		"udp=true",
	}
	if hysteria2.ObfsPassword != "" {
		configs = append(configs, "obfs=salamander", fmt.Sprintf("salamander-password=%s", hysteria2.ObfsPassword))
	}
	if hysteria2.SecurityConfig.SNI != "" {
		configs = append(configs, fmt.Sprintf("sni=%s", hysteria2.SecurityConfig.SNI))
		if hysteria2.SecurityConfig.AllowInsecure {
			configs = append(configs, "skip-cert-verify=true")
		} else {
			configs = append(configs, "skip-cert-verify=false")
		}
	}
	uri := strings.Join(configs, ",")
	return uri + "\r\n"
}
