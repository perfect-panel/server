package singbox

import (
	"encoding/json"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
	"github.com/perfect-panel/server/pkg/logger"
)

func BuildSingbox(adapter proxy.Adapter, uuid string) ([]byte, error) {
	// build outbounds type is Proxy
	var proxies []Proxy
	// build outbound group
	for _, group := range adapter.Group {
		if group.Type == proxy.GroupTypeSelect {
			selector := Proxy{
				Type: Selector,
				Tag:  group.Name,
				SelectorOptions: &SelectorOutboundOptions{
					OutboundOptions: OutboundOptions{
						Tag:  group.Name,
						Type: Selector,
					},
					Outbounds:                 group.Proxies,
					Default:                   group.Proxies[0],
					InterruptExistConnections: false,
				},
			}
			proxies = append(proxies, selector)
		} else if group.Type == proxy.GroupTypeURLTest {
			selector := Proxy{
				Type: URLTest,
				Tag:  group.Name,
				URLTestOptions: &URLTestOutboundOptions{
					OutboundOptions: OutboundOptions{
						Tag:  group.Name,
						Type: URLTest,
					},
					Outbounds: group.Proxies,
					URL:       group.URL,
				},
			}
			proxies = append(proxies, selector)
		} else {
			logger.Errorf("[sing-box] Unknown group type: %s, group name: %s", group.Type, group.Name)
		}
	}

	// build outbounds
	for _, data := range adapter.Proxies {
		p := buildProxy(data, uuid)
		if p == nil {
			continue
		}
		proxies = append(proxies, *p)
	}

	// add direct outbound
	direct := Proxy{
		Type: Direct,
		Tag:  "DIRECT",
	}
	// add block outbound
	block := Proxy{
		Type: Block,
		Tag:  "block",
	}
	// add dns outbound
	dns := Proxy{
		Type: DNS,
		Tag:  "dns-out",
	}
	proxies = append(proxies, direct, block, dns)

	var rawConfig map[string]any
	if err := json.Unmarshal([]byte(DefaultTemplate), &rawConfig); err != nil {
		return nil, err
	}

	rawConfig["outbounds"] = proxies
	route := RouteOptions{
		Final: "手动选择",
		Rules: []Rule{
			{
				Inbound: []string{
					"tun-in",
					"mixed-in",
				},
				Action: "sniff",
			},
			{
				Type: "logical",
				Mode: "or",
				Rules: []Rule{
					{
						Port: []uint16{53},
					},
					{
						Protocol: []string{"dns"},
					},
				},
				Action: "hijack-dns",
			},
			{
				RuleSet: []string{
					"geosite-category-ads-all",
				},
				ClashMode: "rule",
				Action:    "reject",
			},
			{
				ClashMode: "direct",
				Outbound:  "DIRECT",
			},
			{
				ClashMode: "global",
				Outbound:  "手动选择",
			},
			{
				IPIsPrivate: true,
				Outbound:    "DIRECT",
			},
			{
				RuleSet: []string{
					"geosite-private",
				},
				Outbound: "DIRECT",
			},
		},
		RuleSet: []RuleSet{
			{
				Tag:            "geoip-cn",
				Type:           "remote",
				Format:         "binary",
				URL:            "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geoip/cn.srs",
				DownloadDetour: "DIRECT",
			},
			{
				Tag:            "geosite-cn",
				Type:           "remote",
				Format:         "binary",
				URL:            "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/cn.srs",
				DownloadDetour: "DIRECT",
			},
			{
				Tag:            "geosite-private",
				Type:           "remote",
				Format:         "binary",
				URL:            "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/private.srs",
				DownloadDetour: "DIRECT",
			},
			{
				Tag:            "geosite-category-ads-all",
				Type:           "remote",
				Format:         "binary",
				URL:            "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/category-ads-all.srs",
				DownloadDetour: "DIRECT",
			},
			{
				Tag:            "geosite-geolocation-!cn",
				Type:           "remote",
				Format:         "binary",
				URL:            "https://testingcf.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@sing/geo/geosite/geolocation-!cn.srs",
				DownloadDetour: "DIRECT",
			},
		},
		AutoDetectInterface: true,
	}
	route.Rules = append(route.Rules, adapterToSingboxRule(adapter.Rules)...)
	rawConfig["route"] = route
	return json.Marshal(rawConfig)
}

func buildProxy(data proxy.Proxy, uuid string) *Proxy {
	var p *Proxy
	var err error
	switch data.Protocol {
	case VLESS:
		p, err = ParseVless(data, uuid)
	case Shadowsocks:
		p, err = ParseShadowsocks(data, uuid)
	case Trojan:
		p, err = ParseTrojan(data, uuid)
	case VMess:
		p, err = ParseVMess(data, uuid)

	case Hysteria2:
		p, err = ParseHysteria2(data, uuid)

	case TUIC:
		p, err = ParseTUIC(data, uuid)

	default:
		logger.Error("Unknown protocol", logger.Field("protocol", data.Protocol), logger.Field("server", data.Name))
	}
	if err != nil {
		logger.Error("ParseVless", logger.Field("error", err.Error()), logger.Field("server", data.Name), logger.Field("protocol", data.Protocol))
		return nil
	}
	return p
}
