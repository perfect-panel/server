package loon

import (
	"bytes"
	"embed"
	"strings"
	"text/template"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
	"github.com/perfect-panel/server/pkg/logger"
)

//go:embed *.tpl
var configFiles embed.FS

func BuildLoon(servers []proxy.Proxy, uuid string) []byte {
	uri := ""
	nodes := make([]string, 0)
	for _, s := range servers {
		switch s.Protocol {
		case "vmess":
			nodes = append(nodes, s.Name)
			uri += buildVMess(s, uuid)
		case "shadowsocks":
			nodes = append(nodes, s.Name)
			uri += buildShadowsocks(s, uuid)
		case "trojan":
			nodes = append(nodes, s.Name)
			uri += buildTrojan(s, uuid)
		case "vless":
			nodes = append(nodes, s.Name)
			uri += buildVless(s, uuid)
		case "hysteria2":
			nodes = append(nodes, s.Name)
			uri += buildHysteria2(s, uuid)
		default:
			continue
		}
	}
	file, err := configFiles.ReadFile("default.tpl")
	if err != nil {
		logger.Errorf("read default surfboard config error: %v", err.Error())
		return nil
	}
	// replace template
	tpl, err := template.New("default").Parse(string(file))
	if err != nil {
		logger.Errorf("read default surfboard config error: %v", err.Error())
		return nil
	}
	var buf bytes.Buffer
	if err = tpl.Execute(&buf, map[string]interface{}{
		"Proxies": uri,
		"Nodes":   strings.Join(nodes, ","),
	}); err != nil {
		logger.Errorf("Execute Loon template error: %v", err.Error())
		return nil
	}

	return buf.Bytes()
}
