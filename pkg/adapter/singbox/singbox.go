package singbox

import (
	"encoding/json"
	"fmt"
)

const (
	Trojan      = "trojan"
	VLESS       = "vless"
	VMess       = "vmess"
	TUIC        = "tuic"
	Hysteria2   = "hysteria2"
	AnyTLS      = "anytls"
	Shadowsocks = "shadowsocks"
	Selector    = "selector"
	URLTest     = "urltest"
	Direct      = "direct"
	Block       = "block"
	DNS         = "dns"
)

type Proxy struct {
	Tag                string                    `json:"tag,omitempty"`
	Type               string                    `json:"type"`
	ShadowsocksOptions *ShadowsocksOptions       `json:"-"`
	TUICOptions        *TUICOutboundOptions      `json:"-"`
	TrojanOptions      *TrojanOutboundOptions    `json:"-"`
	VLESSOptions       *VLESSOutboundOptions     `json:"-"`
	VMessOptions       *VMessOutboundOptions     `json:"-"`
	AnyTLSOptions      *AnyTLSOutboundOptions    `json:"-"`
	Hysteria2Options   *Hysteria2OutboundOptions `json:"-"`
	SelectorOptions    *SelectorOutboundOptions  `json:"-"`
	URLTestOptions     *URLTestOutboundOptions   `json:"-"`
}

type ServerOptions struct {
	Tag        string `json:"tag"`
	Type       string `json:"type"`
	Server     string `json:"server"`
	ServerPort int    `json:"server_port,omitempty"`
}
type OutboundOptions struct {
	Tag  string `json:"tag"`
	Type string `json:"type"`
}
type SelectorOutboundOptions struct {
	OutboundOptions
	Outbounds                 []string `json:"outbounds"`
	Default                   string   `json:"default,omitempty"`
	InterruptExistConnections bool     `json:"interrupt_exist_connections,omitempty"`
}

type URLTestOutboundOptions struct {
	OutboundOptions
	Outbounds                 []string `json:"outbounds"`
	URL                       string   `json:"url,omitempty"`
	Interval                  Duration `json:"interval,omitempty"`
	Tolerance                 uint16   `json:"tolerance,omitempty"`
	IdleTimeout               Duration `json:"idle_timeout,omitempty"`
	InterruptExistConnections bool     `json:"interrupt_exist_connections,omitempty"`
}

type RouteOptions struct {
	Rules               []Rule    `json:"rules,omitempty"`
	Final               string    `json:"final,omitempty"`
	RuleSet             []RuleSet `json:"rule_set,omitempty"`
	AutoDetectInterface bool      `json:"auto_detect_interface,omitempty"`
}

func (p Proxy) MarshalJSON() ([]byte, error) {
	type Alias Proxy
	aux := struct {
		Alias
	}{
		Alias: (Alias)(p),
	}
	switch p.Type {
	case Shadowsocks:
		return json.Marshal(p.ShadowsocksOptions)
	case TUIC:
		return json.Marshal(p.TUICOptions)
	case Trojan:
		return json.Marshal(p.TrojanOptions)
	case VLESS:
		return json.Marshal(p.VLESSOptions)
	case VMess:
		return json.Marshal(p.VMessOptions)
	case Hysteria2:
		return json.Marshal(p.Hysteria2Options)
	case AnyTLS:
		return json.Marshal(p.AnyTLSOptions)
	case Selector:
		return json.Marshal(p.SelectorOptions)
	case URLTest:
		return json.Marshal(p.URLTestOptions)
	case Direct, Block, DNS:
		return json.Marshal(aux.Alias)
	default:
		return nil, fmt.Errorf("[sing-box] MarshalJSON unknown type: %s", p.Type)
	}
}
