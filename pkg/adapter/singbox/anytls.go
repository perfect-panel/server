package singbox

import "github.com/perfect-panel/server/pkg/adapter/proxy"

type AnyTLSOutboundOptions struct {
	ServerOptions
	OutboundTLSOptionsContainer
	Password string `json:"password,omitempty"`
}

func ParseAnyTLS(data proxy.Proxy, password string) (*Proxy, error) {
	anyTLS := data.Option.(proxy.AnyTLS)

	config := &AnyTLSOutboundOptions{
		ServerOptions: ServerOptions{
			Tag:        data.Name,
			Type:       AnyTLS,
			Server:     data.Server,
			ServerPort: data.Port,
		},
		OutboundTLSOptionsContainer: OutboundTLSOptionsContainer{
			TLS: &OutboundTLSOptions{
				Enabled:  true,
				ALPN:     []string{"h2", "http/1.1"},
				Insecure: anyTLS.SecurityConfig.AllowInsecure,
			},
		},
		Password: password,
	}

	if anyTLS.SecurityConfig.SNI != "" {
		config.OutboundTLSOptionsContainer.TLS.ServerName = anyTLS.SecurityConfig.SNI
	}

	p := &Proxy{
		Tag:           data.Name,
		Type:          AnyTLS,
		AnyTLSOptions: config,
	}

	return p, nil
}
