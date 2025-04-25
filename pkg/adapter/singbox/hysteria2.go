package singbox

import (
	"strings"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

type Hysteria2Obfs struct {
	Type     string `json:"type,omitempty"`
	Password string `json:"password,omitempty"`
}

type Hysteria2OutboundOptions struct {
	ServerOptions
	ServerPorts []string       `json:"server_ports,omitempty"`
	HopInterval int            `json:"hop_interval,omitempty"`
	UpMbps      int            `json:"up_mbps,omitempty"`
	DownMbps    int            `json:"down_mbps,omitempty"`
	Obfs        *Hysteria2Obfs `json:"obfs,omitempty"`
	Password    string         `json:"password,omitempty"`
	Network     string         `json:"network,omitempty"`
	OutboundTLSOptionsContainer
	Multiplex *OutboundMultiplexOptions `json:"multiplex,omitempty"`
	Transport *V2RayTransportOptions    `json:"transport,omitempty"`
}

func ParseHysteria2(data proxy.Proxy, password string) (*Proxy, error) {
	hysteria2 := data.Option.(proxy.Hysteria2)

	p := &Proxy{
		Tag:  data.Name,
		Type: Hysteria2,
		Hysteria2Options: &Hysteria2OutboundOptions{
			ServerOptions: ServerOptions{
				Tag:    data.Name,
				Type:   Hysteria2,
				Server: data.Server,
			},
			Password: password,
		},
	}

	var ports []string

	if hysteria2.HopPorts != "" {
		ps := strings.Split(hysteria2.HopPorts, ",")
		for _, port := range ps {
			// 舍弃单个端口，只保留端口范围
			if len(strings.Split(port, "-")) > 1 {
				tmp := strings.Split(port, "-")
				ports = append(ports, strings.Join(tmp, ":"))
			}
		}

	}
	if len(ports) > 0 {
		p.Hysteria2Options.ServerPorts = ports
		p.Hysteria2Options.HopInterval = hysteria2.HopInterval
	} else {
		p.Hysteria2Options.ServerPort = data.Port
	}

	if hysteria2.ObfsPassword != "" {
		p.Hysteria2Options.Obfs = &Hysteria2Obfs{
			Type:     "salamander",
			Password: hysteria2.ObfsPassword,
		}
	}
	var tls *OutboundTLSOptions
	if hysteria2.SecurityConfig.SNI != "" {
		tls = NewOutboundTLSOptions("tls", hysteria2.SecurityConfig)
	}
	p.Hysteria2Options.TLS = tls
	return p, nil
}
