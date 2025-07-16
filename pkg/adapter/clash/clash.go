package clash

import (
	"fmt"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
	"github.com/perfect-panel/server/pkg/logger"
	"gopkg.in/yaml.v3"
)

type Clash struct {
	proxy.Adapter
}

func NewClash(adapter proxy.Adapter) *Clash {
	return &Clash{
		Adapter: adapter,
	}
}

func (c *Clash) Build(uuid string) ([]byte, error) {
	var proxies []Proxy
	for _, v := range c.Proxies {
		p, err := c.parseProxy(v, uuid)
		if err != nil {
			logger.Errorf("Failed to parse proxy for %s: %s", v.Name, err.Error())
			continue
		}
		proxies = append(proxies, *p)
	}
	var rawConfig RawConfig
	if err := yaml.Unmarshal([]byte(DefaultTemplate), &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}
	rawConfig.Proxies = proxies
	// generate proxy groups
	var groups []ProxyGroup
	for _, group := range c.Group {
		groups = append(groups, ProxyGroup{
			Name:     group.Name,
			Type:     string(group.Type),
			Proxies:  group.Proxies,
			Url:      group.URL,
			Interval: group.Interval,
		})
	}
	rawConfig.ProxyGroups = groups
	rawConfig.Rules = append(c.Rules, fmt.Sprintf("MATCH,%s", c.Default))
	return yaml.Marshal(&rawConfig)
}

func (c *Clash) parseProxy(p proxy.Proxy, uuid string) (*Proxy, error) {
	parseFuncs := map[string]func(proxy.Proxy, string) (*Proxy, error){
		"shadowsocks": parseShadowsocks,
		"trojan":      parseTrojan,
		"vless":       parseVless,
		"vmess":       parseVmess,
		"hysteria2":   parseHysteria2,
		"tuic":        parseTuic,
		"anytls":      parseAnyTLS,
	}

	if parseFunc, exists := parseFuncs[p.Protocol]; exists {
		return parseFunc(p, uuid)
	}

	logger.Errorw("Unknown protocol", logger.Field("protocol", p.Protocol), logger.Field("server", p.Name))
	return nil, fmt.Errorf("unknown protocol: %s", p.Protocol)
}
