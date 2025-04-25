package quantumultx

import (
	"fmt"
	"strings"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

func buildShadowsocks(data proxy.Proxy, uuid string) string {
	ss := data.Option.(proxy.Shadowsocks)
	addr := fmt.Sprintf("%s:%d", data.Server, data.Port)

	config := []string{
		addr,
		fmt.Sprintf("method=%s", ss.Method),
		fmt.Sprintf("password=%s", uuid),
		"fast-open=true",
		"udp-relay=true",
		fmt.Sprintf("tag=%s", data.Name),
	}
	return strings.Join(config, ",") + "\r\n"
}
