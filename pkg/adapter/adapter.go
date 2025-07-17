package adapter

import (
	"embed"

	"github.com/perfect-panel/server/internal/model/server"
	"github.com/perfect-panel/server/pkg/adapter/clash"
	"github.com/perfect-panel/server/pkg/adapter/general"
	"github.com/perfect-panel/server/pkg/adapter/loon"
	"github.com/perfect-panel/server/pkg/adapter/proxy"
	"github.com/perfect-panel/server/pkg/adapter/quantumultx"
	"github.com/perfect-panel/server/pkg/adapter/shadowrocket"
	"github.com/perfect-panel/server/pkg/adapter/singbox"
	"github.com/perfect-panel/server/pkg/adapter/surfboard"
	"github.com/perfect-panel/server/pkg/adapter/v2rayn"
)

//go:embed template/*
var TemplateFS embed.FS

var (
	AutoSelect = "Auto - UrlTest"
)

type Config struct {
	Nodes []*server.Server
	Rules []*server.RuleGroup
	Tags  map[string][]*server.Server
}

type Adapter struct {
	proxy.Adapter
}

func NewAdapter(cfg *Config) *Adapter {
	// 转换服务器列表
	proxies, nodes, tags := adapterProxies(cfg.Nodes)
	// 转换规则组
	g, r, d := adapterRules(cfg.Rules)
	if d == "" {
		d = AutoSelect
	}
	// 生成默认代理组
	proxyGroup := append(generateDefaultGroup(), g...)
	// 合并代理组
	proxyGroup = SortGroups(proxyGroup, nodes, tags, d)
	return &Adapter{
		Adapter: proxy.Adapter{
			Proxies:    proxies,
			Group:      proxyGroup,
			Rules:      r,
			Nodes:      nodes,
			Default:    d,
			TemplateFS: &TemplateFS,
		},
	}
}

// BuildClash generates a Clash configuration for the given UUID.
func (m *Adapter) BuildClash(uuid string) ([]byte, error) {
	client := clash.NewClash(m.Adapter)
	return client.Build(uuid)
}

// BuildGeneral generates a general configuration for the given UUID.
func (m *Adapter) BuildGeneral(uuid string) []byte {
	return general.GenerateBase64General(m.Proxies, uuid)
}

// BuildLoon generates a Loon configuration for the given UUID.
func (m *Adapter) BuildLoon(uuid string) []byte {
	return loon.BuildLoon(m.Proxies, uuid)
}

// BuildQuantumultX generates a Quantumult X configuration for the given UUID.
func (m *Adapter) BuildQuantumultX(uuid string) string {
	return quantumultx.BuildQuantumultX(m.Proxies, uuid)
}

// BuildSingbox generates a Singbox configuration for the given UUID.
func (m *Adapter) BuildSingbox(uuid string) ([]byte, error) {
	return singbox.BuildSingbox(m.Adapter, uuid)
}
func (m *Adapter) BuildShadowrocket(uuid string, userInfo shadowrocket.UserInfo) []byte {
	return shadowrocket.BuildShadowrocket(m.Proxies, uuid, userInfo)
}

// BuildSurfboard generates a Surfboard configuration for the given site name and user info.
func (m *Adapter) BuildSurfboard(siteName string, user surfboard.UserInfo) []byte {
	return surfboard.BuildSurfboard(m.Adapter, siteName, user)
}

// BuildV2rayN generates a V2rayN configuration for the given UUID.
func (m *Adapter) BuildV2rayN(uuid string) []byte {
	return v2rayn.NewV2rayN(m.Adapter).Build(uuid)
}
