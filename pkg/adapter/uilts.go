package adapter

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/pkg/adapter/proxy"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/random"
	"github.com/perfect-panel/server/pkg/tool"
)

// addNode creates a new proxy node based on the provided server data and host/port.
func addNode(data *server.Server, host string, port int) *proxy.Proxy {
	var option any
	tags := strings.Split(data.Tags, ",")
	if len(tags) > 0 {
		tags = tool.RemoveDuplicateElements(tags...)
	}

	node := proxy.Proxy{
		Name:     data.Name,
		Server:   host,
		Port:     port,
		Country:  data.Country,
		Protocol: data.Protocol,
		Tags:     tags,
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

func adapterRules(groups []*server.RuleGroup) (proxyGroup []proxy.Group, rules []string, defaultGroup string) {
	for _, group := range groups {
		if group.Default {
			log.Printf("[Debug] 规则组 %s 是默认组", group.Name)
			defaultGroup = group.Name
		}
		switch group.Type {
		case server.RuleGroupTypeReject:
			proxyGroup = append(proxyGroup, proxy.Group{
				Name:    group.Name,
				Type:    proxy.GroupTypeSelect,
				Proxies: []string{"REJECT", "DIRECT", AutoSelect},
				Reject:  true,
			})
		case server.RuleGroupTypeDirect:
			proxyGroup = append(proxyGroup, proxy.Group{
				Name:    group.Name,
				Type:    proxy.GroupTypeSelect,
				Proxies: []string{"DIRECT", AutoSelect},
				Direct:  true,
			})
		default:
			proxyGroup = append(proxyGroup, proxy.Group{
				Name:    group.Name,
				Type:    proxy.GroupTypeSelect,
				Proxies: []string{},
				Tags:    RemoveEmptyString(strings.Split(group.Tags, ",")),
				Default: group.Default,
			})
		}

		rules = append(rules, strings.Split(group.Rules, "\n")...)
	}
	log.Printf("[Dapter] 生成规则组: %d", len(proxyGroup))
	return proxyGroup, tool.RemoveDuplicateElements(rules...), defaultGroup
}

// generateDefaultGroup generates a default proxy group with auto-selection and manual selection options.
func generateDefaultGroup() (proxyGroup []proxy.Group) {
	proxyGroup = append(proxyGroup, proxy.Group{
		Name:     AutoSelect,
		Type:     proxy.GroupTypeURLTest,
		Proxies:  make([]string, 0),
		URL:      "https://www.gstatic.com/generate_204",
		Interval: 300,
	})

	return proxyGroup
}

func adapterProxies(servers []*server.Server) ([]proxy.Proxy, []string, map[string][]string) {
	var proxies []proxy.Proxy
	var tags = make(map[string][]string)
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
					n.Name = relay.Prefix + n.Name
				}
				if node.Tags != "" {
					t := tool.RemoveDuplicateElements(strings.Split(node.Tags, ",")...)
					for _, tag := range t {
						if tag != "" {
							if _, ok := tags[tag]; !ok {
								tags[tag] = []string{}
							}
							tags[tag] = append(tags[tag], n.Name)
						}
					}
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
				n.Name = relay.Prefix + node.Name
			}
			if node.Tags != "" {
				t := tool.RemoveDuplicateElements(strings.Split(node.Tags, ",")...)
				for _, tag := range t {
					if tag != "" {
						if _, ok := tags[tag]; !ok {
							tags[tag] = []string{}
						}
						tags[tag] = append(tags[tag], n.Name)
					}
				}
			}
			proxies = append(proxies, *n)
		default:
			logger.Info("Not Relay Mode", logger.Field("node", node.Name), logger.Field("relayMode", node.RelayMode))
			n := addNode(node, node.ServerAddr, 0)
			if n != nil {
				if node.Tags != "" {
					t := tool.RemoveDuplicateElements(strings.Split(node.Tags, ",")...)
					for _, tag := range t {
						if tag != "" {
							if _, ok := tags[tag]; !ok {
								tags[tag] = []string{}
							}
							tags[tag] = append(tags[tag], n.Name)
						}
					}
				}
				proxies = append(proxies, *n)
			}
		}
	}

	var nodes []string
	for _, p := range proxies {
		nodes = append(nodes, p.Name)
	}

	return proxies, tool.RemoveDuplicateElements(nodes...), tags
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

// SortGroups sorts the provided slice of proxy groups by their names.
func SortGroups(groups []proxy.Group, nodes []string, tags map[string][]string, defaultName string) []proxy.Group {
	var sortedGroups []proxy.Group
	var defaultGroup, autoSelectGroup proxy.Group
	// 在所有分组找到默认分组并将他放到第一个
	for _, group := range groups {
		if group.Name == "" || group.Name == "DIRECT" || group.Name == "REJECT" {
			continue
		}
		// 如果是默认分组
		if group.Default {
			group.Proxies = append([]string{AutoSelect}, nodes...)
			group.Proxies = append(group.Proxies, "DIRECT")
			defaultGroup = group
			continue
		}
		if group.Reject || group.Direct {
			if defaultName != AutoSelect {
				group.Proxies = append(group.Proxies, defaultName)
			}
			sortedGroups = append(sortedGroups, group)
			continue
		}

		if group.Name == AutoSelect {
			group.Proxies = nodes
			autoSelectGroup = group
			continue
		}
		// Tags 分组
		if len(group.Tags) > 0 {
			var proxies []string
			for _, tag := range group.Tags {
				if node, ok := tags[tag]; ok {
					proxies = append(proxies, node...)
				}
			}
			group.Proxies = append(tool.RemoveDuplicateElements(proxies...), AutoSelect, "DIRECT")
			sortedGroups = append(sortedGroups, group)
			continue
		}

		group.Proxies = append([]string{AutoSelect}, nodes...)
		group.Proxies = append(group.Proxies, "DIRECT")
		group.Proxies = tool.RemoveElementBySlice(group.Proxies, group.Name)
		sortedGroups = append(sortedGroups, group)
	}

	if defaultGroup.Name != "" {
		sortedGroups = append([]proxy.Group{defaultGroup}, sortedGroups...)
	}
	if autoSelectGroup.Name != "" && autoSelectGroup.Name != defaultGroup.Name {
		sortedGroups = append(sortedGroups, autoSelectGroup)
	}

	return sortedGroups

}
