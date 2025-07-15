package adapter

import (
	"encoding/json"
	"strings"

	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/pkg/adapter/proxy"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/random"
	"github.com/perfect-panel/server/pkg/tool"
)

func addNode(data *server.Server, host string, port int) *proxy.Proxy {
	var option any
	node := proxy.Proxy{
		Name:     data.Name,
		Server:   host,
		Port:     port,
		Country:  data.Country,
		Protocol: data.Protocol,
	}
	switch data.Protocol {
	case "shadowsocks":
		var ss proxy.Shadowsocks
		if err := json.Unmarshal([]byte(data.Config), &ss); err != nil {
			return nil
		}
		if port == 0 {
			node.Port = ss.Port
		}
		option = ss
	case "vless":
		var vless proxy.Vless
		if err := json.Unmarshal([]byte(data.Config), &vless); err != nil {
			return nil
		}
		if port == 0 {
			node.Port = vless.Port
		}
		option = vless
	case "vmess":
		var vmess proxy.Vmess
		if err := json.Unmarshal([]byte(data.Config), &vmess); err != nil {
			return nil
		}
		if port == 0 {
			node.Port = vmess.Port
		}
		option = vmess
	case "trojan":
		var trojan proxy.Trojan
		if err := json.Unmarshal([]byte(data.Config), &trojan); err != nil {
			return nil
		}
		if port == 0 {
			node.Port = trojan.Port
		}
		option = trojan
	case "hysteria2":
		var hysteria2 proxy.Hysteria2
		if err := json.Unmarshal([]byte(data.Config), &hysteria2); err != nil {
			return nil
		}
		if port == 0 {
			node.Port = hysteria2.Port
		}
		option = hysteria2
	case "tuic":
		var tuic proxy.Tuic
		if err := json.Unmarshal([]byte(data.Config), &tuic); err != nil {
			return nil
		}
		if port == 0 {
			node.Port = tuic.Port
		}
		option = tuic
	default:
		return nil
	}
	node.Option = option
	return &node
}

func addProxyToGroup(proxyName, groupName string, groups []proxy.Group) []proxy.Group {
	for i, group := range groups {
		if group.Name == groupName {
			groups[i].Proxies = tool.RemoveDuplicateElements(append(group.Proxies, proxyName)...)
			return groups
		}
	}
	groups = append(groups, proxy.Group{
		Name:    groupName,
		Type:    proxy.GroupTypeSelect,
		Proxies: []string{proxyName},
	})
	return groups
}

func adapterRules(groups []*server.RuleGroup) (proxyGroup []proxy.Group, rules []string) {
	for _, group := range groups {
		proxyGroup = append(proxyGroup, proxy.Group{
			Name:    group.Name,
			Type:    proxy.GroupTypeSelect,
			Proxies: RemoveEmptyString(strings.Split(group.Tags, ",")),
		})
		rules = append(rules, strings.Split(group.Rules, "\n")...)
	}
	return
}

func adapterTags(tags map[string][]*server.Server, group []proxy.Group) (proxyGroup []proxy.Group) {
	for tag, servers := range tags {
		proxies := adapterProxies(servers)
		if len(proxies) != 0 {
			for _, p := range proxies {
				group = addProxyToGroup(p.Name, tag, group)
			}
		}
	}
	return group
}

func generateProxyGroup(servers []proxy.Proxy) (proxyGroup []proxy.Group, nodes []string) {
	// 设置手动选择分组
	proxyGroup = append(proxyGroup, []proxy.Group{
		{
			Name:     "智能线路",
			Type:     proxy.GroupTypeURLTest,
			Proxies:  make([]string, 0),
			URL:      "https://www.gstatic.com/generate_204",
			Interval: 300,
		},
		{
			Name:    "手动选择",
			Type:    proxy.GroupTypeSelect,
			Proxies: []string{"智能线路"},
		},
	}...)

	for _, node := range servers {
		proxyGroup = addProxyToGroup(node.Name, "智能线路", proxyGroup)
		proxyGroup = addProxyToGroup(node.Name, "手动选择", proxyGroup)
		nodes = append(nodes, node.Name)
	}

	proxyGroup = addProxyToGroup("DIRECT", "手动选择", proxyGroup)
	return proxyGroup, tool.RemoveDuplicateElements(nodes...)
}

func adapterProxies(servers []*server.Server) []proxy.Proxy {
	var proxies []proxy.Proxy
	for _, node := range servers {
		switch node.RelayMode {
		case server.RelayModeAll:
			var relays []server.NodeRelay
			if err := json.Unmarshal([]byte(node.RelayNode), &relays); err != nil {
				logger.Errorw("Unmarshal RelayNode", logger.Field("error", err.Error()), logger.Field("node", node.Name), logger.Field("relayNode", node.RelayNode))
				continue
			}
			for _, relay := range relays {
				n := addNode(node, relay.Host, relay.Port)
				if n == nil {
					continue
				}
				if relay.Prefix != "" {
					n.Name = relay.Prefix + "-" + n.Name
				}
				proxies = append(proxies, *n)
			}
		case server.RelayModeRandom:
			var relays []server.NodeRelay
			if err := json.Unmarshal([]byte(node.RelayNode), &relays); err != nil {
				logger.Errorw("Unmarshal RelayNode", logger.Field("error", err.Error()), logger.Field("node", node.Name), logger.Field("relayNode", node.RelayNode))
				continue
			}
			randNum := random.RandomInRange(0, len(relays)-1)
			relay := relays[randNum]
			n := addNode(node, relay.Host, relay.Port)
			if n == nil {
				continue
			}
			if relay.Prefix != "" {
				n.Name = relay.Prefix + " - " + node.Name
			}
			proxies = append(proxies, *n)
		default:
			logger.Info("Not Relay Mode", logger.Field("node", node.Name), logger.Field("relayMode", node.RelayMode))
			n := addNode(node, node.ServerAddr, 0)
			if n != nil {
				proxies = append(proxies, *n)
			}
		}
	}
	return proxies
}

// RemoveEmptyString 切片去除空值
func RemoveEmptyString(arr []string) []string {
	var result []string
	for _, str := range arr {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}

func RemoveEmptyGroup(arr []proxy.Group) []proxy.Group {
	var result []proxy.Group
	var removeNames []string
	for _, group := range arr {
		if group.Name == "手动选择" {
			group.Proxies = tool.RemoveStringElement(group.Proxies, removeNames...)
		}
		if len(group.Proxies) > 0 {
			result = append(result, group)
		} else {
			removeNames = append(removeNames, group.Name)
		}
	}
	return result
}
