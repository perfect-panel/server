package singbox

import (
	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

type VLESSOutboundOptions struct {
	ServerOptions
	OutboundTLSOptionsContainer
	UUID           string                    `json:"uuid"`
	Flow           string                    `json:"flow,omitempty"`
	Network        string                    `json:"network,omitempty"`
	Multiplex      *OutboundMultiplexOptions `json:"multiplex,omitempty"`
	Transport      *V2RayTransportOptions    `json:"transport,omitempty"`
	PacketEncoding *string                   `json:"packet_encoding,omitempty"`
}

func ParseVless(data proxy.Proxy, uuid string) (*Proxy, error) {
	vless := data.Option.(proxy.Vless)
	packetEncoding := "xudp"
	p := &Proxy{
		Tag:  data.Name,
		Type: VLESS,
		VLESSOptions: &VLESSOutboundOptions{
			ServerOptions: ServerOptions{
				Tag:        data.Name,
				Type:       VLESS,
				Server:     data.Server,
				ServerPort: data.Port,
			},
			UUID:           uuid,
			Flow:           vless.Flow,
			PacketEncoding: &packetEncoding,
		},
	}
	// Transport options
	transport := NewV2RayTransportOptions(vless.Transport, vless.TransportConfig)
	p.VLESSOptions.Transport = transport

	// Security options
	p.VLESSOptions.TLS = NewOutboundTLSOptions(vless.Security, vless.SecurityConfig)

	return p, nil
}
