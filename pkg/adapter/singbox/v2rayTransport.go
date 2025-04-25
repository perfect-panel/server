package singbox

import (
	"encoding/json"
	"time"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

type V2RayTransportOptions struct {
	Type               string                  `json:"type"`
	HTTPOptions        V2RayHTTPOptions        `json:"-"`
	WebsocketOptions   V2RayWebsocketOptions   `json:"-"`
	QUICOptions        V2RayQUICOptions        `json:"-"`
	GRPCOptions        V2RayGRPCOptions        `json:"-"`
	HTTPUpgradeOptions V2RayHTTPUpgradeOptions `json:"-"`
}

func (v V2RayTransportOptions) MarshalJSON() ([]byte, error) {
	var v2rayTransportOptions any
	data := map[string]any{
		"type": v.Type,
	}
	switch v.Type {
	case "http":
		v2rayTransportOptions = v.HTTPOptions
	case "ws":
		v2rayTransportOptions = v.WebsocketOptions
	case "quic":
		v2rayTransportOptions = v.QUICOptions
	case "grpc":
		v2rayTransportOptions = v.GRPCOptions
	case "httpupgrade":
		v2rayTransportOptions = v.HTTPUpgradeOptions
	}
	if err := mergeOptions(data, v2rayTransportOptions); err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func NewV2RayTransportOptions(network string, transport proxy.TransportConfig) *V2RayTransportOptions {
	var t *V2RayTransportOptions = nil
	switch network {
	case "websocket":
		t = &V2RayTransportOptions{
			Type: "ws",
			WebsocketOptions: V2RayWebsocketOptions{
				Path: transport.Path,
				Headers: map[string]Listable[string]{
					"Host": []string{transport.Host},
				},
				MaxEarlyData:        2048,
				EarlyDataHeaderName: "Sec-WebSocket-Protocol",
			},
		}
	case "httpupgrade":
		t = &V2RayTransportOptions{
			Type: "httpupgrade",
			HTTPOptions: V2RayHTTPOptions{
				Path: transport.Path,
				Host: []string{transport.Host},
				Headers: map[string]Listable[string]{
					"Host": []string{transport.Host},
				},
			},
		}

	case "grpc":
		t = &V2RayTransportOptions{
			Type: "grpc",
			GRPCOptions: V2RayGRPCOptions{
				ServiceName: transport.ServiceName,
			},
		}
	}
	return t
}

type V2RayHTTPOptions struct {
	Host        Listable[string] `json:"host,omitempty"`
	Path        string           `json:"path,omitempty"`
	Method      string           `json:"method,omitempty"`
	Headers     HTTPHeader       `json:"headers,omitempty"`
	IdleTimeout Duration         `json:"idle_timeout,omitempty"`
	PingTimeout Duration         `json:"ping_timeout,omitempty"`
}

type V2RayWebsocketOptions struct {
	Path                string     `json:"path,omitempty"`
	Headers             HTTPHeader `json:"headers,omitempty"`
	MaxEarlyData        uint32     `json:"max_early_data,omitempty"`
	EarlyDataHeaderName string     `json:"early_data_header_name,omitempty"`
}

type V2RayQUICOptions struct{}

type V2RayGRPCOptions struct {
	ServiceName         string `json:"service_name,omitempty"`
	IdleTimeout         string `json:"idle_timeout,omitempty"`
	PingTimeout         string `json:"ping_timeout,omitempty"`
	PermitWithoutStream bool   `json:"permit_without_stream,omitempty"`
	ForceLite           bool   `json:"-"` // for test
}

type V2RayHTTPUpgradeOptions struct {
	Host    string     `json:"host,omitempty"`
	Path    string     `json:"path,omitempty"`
	Headers HTTPHeader `json:"headers,omitempty"`
}

type HTTPHeader map[string]Listable[string]

type Duration time.Duration
