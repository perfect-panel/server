package singbox

import (
	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

type VMessOutboundOptions struct {
	ServerOptions
	UUID                string                    `json:"uuid"`
	Security            string                    `json:"security"`
	AlterId             int                       `json:"alter_id,omitempty"`
	GlobalPadding       bool                      `json:"global_padding,omitempty"`
	AuthenticatedLength bool                      `json:"authenticated_length,omitempty"`
	Network             string                    `json:"network,omitempty"`
	PacketEncoding      string                    `json:"packet_encoding,omitempty"`
	Multiplex           *OutboundMultiplexOptions `json:"multiplex,omitempty"`
	Transport           *V2RayTransportOptions    `json:"transport,omitempty"`
	OutboundTLSOptionsContainer
}

func ParseVMess(data proxy.Proxy, uuid string) (*Proxy, error) {
	vmess := data.Option.(proxy.Vmess)
	p := &Proxy{
		Type: VMess,
		VMessOptions: &VMessOutboundOptions{
			ServerOptions: ServerOptions{
				Tag:        data.Name,
				Type:       VMess,
				Server:     data.Server,
				ServerPort: data.Port,
			},
			UUID:     uuid,
			Security: "auto",
			AlterId:  0,
		},
	}
	// Transport options
	p.VMessOptions.Transport = NewV2RayTransportOptions(vmess.Transport, vmess.TransportConfig)
	// Security options
	p.VMessOptions.TLS = NewOutboundTLSOptions(vmess.Security, vmess.SecurityConfig)

	return p, nil
}
