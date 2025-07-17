package clash

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
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
	for _, proxied := range c.Adapter.Proxies {
		p, err := c.parseProxy(proxied, uuid)
		if err != nil {
			logger.Errorw("Failed to parse proxy", logger.Field("error", err), logger.Field("proxy", p.Name))
			continue
		}
		proxies = append(proxies, *p)
	}
	var groups []ProxyGroup
	for _, group := range c.Adapter.Group {
		groups = append(groups, ProxyGroup{
			Name:     group.Name,
			Type:     string(group.Type),
			Proxies:  group.Proxies,
			Url:      group.URL,
			Interval: group.Interval,
		})
	}
	var rules = append(c.Rules, fmt.Sprintf("MATCH,%s", c.Default))

	tmplBytes, err := c.TemplateFS.ReadFile("template/clash.tpl")
	if err != nil {
		logger.Errorw("Failed to read template file", logger.Field("error", err))
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}
	tpl, err := template.New("clash.yaml").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
		"toYaml": func(v interface{}) string {
			out, err := yaml.Marshal(v)
			if err != nil {
				return fmt.Sprintf("# YAML encode error: %v", err.Error())
			}
			return string(out)
		},
	}).Parse(string(tmplBytes))
	if err != nil {
		logger.Errorw("[Clash] Failed to parse template", logger.Field("error", err))
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]interface{}{
		"Proxies":     proxies,
		"ProxyGroups": groups,
		"Rules":       rules,
	})
	if err != nil {
		logger.Errorw("[Clash] Failed to execute template", logger.Field("error", err))
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
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
