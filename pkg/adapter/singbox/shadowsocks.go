package singbox

import (
	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

type ShadowsocksOptions struct {
	ServerOptions
	Method        string `json:"method,omitempty"`
	Password      string `json:"password,omitempty"`
	Plugin        string `json:"plugin,omitempty"`
	PluginOptions string `json:"plugin_opts,omitempty"`
	Network       string `json:"network,omitempty"`
}

func ParseShadowsocks(data proxy.Proxy, uuid string) (*Proxy, error) {
	config := data.Option.(proxy.Shadowsocks)
	p := &Proxy{
		Tag:  data.Name,
		Type: Shadowsocks,
		ShadowsocksOptions: &ShadowsocksOptions{
			ServerOptions: ServerOptions{
				Tag:        data.Name,
				Type:       Shadowsocks,
				Server:     data.Server,
				ServerPort: data.Port,
			},
			Method:   config.Method,
			Password: uuid,
			Network:  "tcp",
		},
	}
	return p, nil
}
