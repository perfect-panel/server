package shadowrocket

import (
	"fmt"
	"strings"

	"encoding/base64"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

func buildVmess(data proxy.Proxy, uuid string) string {
	vmess := data.Option.(proxy.Vmess)

	userinfo := fmt.Sprintf("auto:%s@%s:%d", uuid, data.Server, data.Port)
	// 准备 config，使用默认值
	config := map[string]interface{}{
		"tfo":     1,
		"remark":  data.Name,
		"alterId": 0,
	}

	// tls 配置
	if vmess.Security == "tls" {
		config["tls"] = 1
		if vmess.SecurityConfig.AllowInsecure {
			config["allowInsecure"] = 1
		}
		if vmess.SecurityConfig.SNI != "" {
			config["peer"] = vmess.SecurityConfig.SNI
		}
	}

	// transport 配置
	switch vmess.Transport {
	case "websocket":
		config["obfs"] = "websocket"
		if vmess.TransportConfig.Path != "" {
			config["path"] = vmess.TransportConfig.Path
		}
		if vmess.TransportConfig.Host != "" {
			config["obfsParam"] = vmess.TransportConfig.Host
		}
	case "grpc":
		config["obfs"] = "grpc"
		if vmess.TransportConfig.ServiceName != "" {
			config["path"] = vmess.TransportConfig.ServiceName
		}
	}
	query := make([]string, 0)
	for k, v := range config {
		query = append(query, fmt.Sprintf("%s=%v", k, v))
	}
	queryStr := strings.Join(query, "&")
	uri := fmt.Sprintf("vmess://%s?%s\r\n", base64.StdEncoding.EncodeToString([]byte(userinfo)), queryStr)
	return uri
}
