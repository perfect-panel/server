package singbox

import (
	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

type TUICOutboundOptions struct {
	ServerOptions
	UUID              string `json:"uuid,omitempty"`
	Password          string `json:"password,omitempty"`
	CongestionControl string `json:"congestion_control,omitempty"`
	UDPRelayMode      string `json:"udp_relay_mode,omitempty"`
	UDPOverStream     bool   `json:"udp_over_stream,omitempty"`
	ZeroRTTHandshake  bool   `json:"zero_rtt_handshake,omitempty"`
	Heartbeat         string `json:"heartbeat,omitempty"`
	Network           string `json:"network,omitempty"`
	OutboundTLSOptionsContainer
}

func ParseTUIC(data proxy.Proxy, uuid string) (*Proxy, error) {
	tuic := data.Option.(proxy.Tuic)
	p := &Proxy{
		Tag:  data.Name,
		Type: TUIC,
		TUICOptions: &TUICOutboundOptions{
			ServerOptions: ServerOptions{
				Tag:        data.Name,
				Type:       TUIC,
				Server:     data.Server,
				ServerPort: data.Port,
			},
			UUID:              uuid,
			Password:          uuid,
			CongestionControl: "bbr",
		},
	}
	// Security options
	p.TUICOptions.TLS = NewOutboundTLSOptions("tls", tuic.SecurityConfig)
	return p, nil
}
