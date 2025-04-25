package singbox

import (
	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

type TrojanOutboundOptions struct {
	ServerOptions
	Password string `json:"password"`
	Network  string `json:"network,omitempty"`
	OutboundTLSOptionsContainer
	Multiplex *OutboundMultiplexOptions `json:"multiplex,omitempty"`
	Transport *V2RayTransportOptions    `json:"transport,omitempty"`
}

func ParseTrojan(data proxy.Proxy, uuid string) (*Proxy, error) {
	trojan := data.Option.(proxy.Trojan)
	p := &Proxy{
		Tag:  data.Name,
		Type: Trojan,
		TrojanOptions: &TrojanOutboundOptions{
			ServerOptions: ServerOptions{
				Tag:        data.Name,
				Type:       Trojan,
				Server:     data.Server,
				ServerPort: data.Port,
			},
			Password: uuid,
		},
	}
	// Transport options
	transport := NewV2RayTransportOptions(trojan.Transport, trojan.TransportConfig)

	p.TrojanOptions.Transport = transport
	// Security options
	p.TrojanOptions.TLS = NewOutboundTLSOptions(trojan.Security, trojan.SecurityConfig)
	return p, nil

}
