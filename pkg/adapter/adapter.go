package adapter

import (
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
	proxies := adapterProxies(cfg.Nodes)
	// 生成代理组
	proxyGroup, nodes := generateProxyGroup(proxies)

	// 转换规则组
	g, r := adapterRules(cfg.Rules)

	// 加入兜底节点
	for i, group := range g {
		if len(group.Proxies) == 0 {
			p := append([]string{"智能线路"}, nodes...)
			g[i].Proxies = append(p, "REJECT", "DIRECT")
		}
	}

	// 合并代理组
	proxyGroup = RemoveEmptyGroup(append(proxyGroup, g...))
	// 处理标签
	proxyGroup = adapterTags(cfg.Tags, proxyGroup)
	return &Adapter{
		Adapter: proxy.Adapter{
			Proxies: proxies,
			Group:   proxyGroup,
			Rules:   r,
			Nodes:   nodes,
		},
	}
}

func (m *Adapter) BuildClash(uuid string) ([]byte, error) {
	client := clash.NewClash(m.Adapter)
	return client.Build(uuid)
}

func (m *Adapter) BuildGeneral(uuid string) []byte {
	return general.GenerateBase64General(m.Proxies, uuid)
}

func (m *Adapter) BuildLoon(uuid string) []byte {
	return loon.BuildLoon(m.Proxies, uuid)
}

func (m *Adapter) BuildQuantumultX(uuid string) string {
	return quantumultx.BuildQuantumultX(m.Proxies, uuid)
}

func (m *Adapter) BuildSingbox(uuid string) ([]byte, error) {
	return singbox.BuildSingbox(m.Adapter, uuid)
}

func (m *Adapter) BuildShadowrocket(uuid string, userInfo shadowrocket.UserInfo) []byte {
	return shadowrocket.BuildShadowrocket(m.Proxies, uuid, userInfo)
}

func (m *Adapter) BuildSurfboard(siteName string, user surfboard.UserInfo) []byte {
	return surfboard.BuildSurfboard(m.Adapter, siteName, user)
}
func (m *Adapter) BuildV2rayN(uuid string) []byte {
	return v2rayn.NewV2rayN(m.Adapter).Build(uuid)
}
