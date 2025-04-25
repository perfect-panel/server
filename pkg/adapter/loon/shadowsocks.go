package loon

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/uuidx"
)

func buildShadowsocks(data proxy.Proxy, password string) string {
	shadowsocks := data.Option.(proxy.Shadowsocks)
	// If the method is 2022-blake3-chacha20-poly1305, it means that the server is a relay server
	if shadowsocks.Method == "2022-blake3-chacha20-poly1305" {
		return ""
	}

	if strings.Contains(shadowsocks.Method, "2022") {
		serverKey, userKey := generateShadowsocks2022Password(shadowsocks, password)
		password = fmt.Sprintf("%s:%s", serverKey, userKey)
	}

	configs := []string{
		fmt.Sprintf("%s=Shadowsocks", data.Name),
		data.Server,
		strconv.Itoa(data.Port),
		shadowsocks.Method,
		password,
		"fast-open=false",
		"udp=true",
	}
	uri := strings.Join(configs, ",")
	return uri + "\r\n"
}

func generateShadowsocks2022Password(ss proxy.Shadowsocks, password string) (string, string) {
	// server key
	var serverKey string
	if ss.Method == "2022-blake3-aes-128-gcm" {
		serverKey = tool.GenerateCipher(ss.ServerKey, 16)
		password = uuidx.UUIDToBase64(password, 16)
	} else {
		serverKey = tool.GenerateCipher(ss.ServerKey, 32)
		password = uuidx.UUIDToBase64(password, 32)
	}
	return serverKey, password
}
