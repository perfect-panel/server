package clash

import (
	"fmt"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

func parseShadowsocks(s proxy.Proxy, uuid string) (*Proxy, error) {
	config, ok := s.Option.(proxy.Shadowsocks)
	if !ok {
		return nil, fmt.Errorf("invalid type for Shadowsocks")
	}
	p := &Proxy{
		Name:     s.Name,
		Type:     "ss",
		Server:   s.Server,
		Port:     s.Port,
		Cipher:   config.Method,
		Password: uuid,
		UDP:      true,
	}

	return p, nil
}

func parseTrojan(data proxy.Proxy, password string) (*Proxy, error) {
	trojan, ok := data.Option.(proxy.Trojan)
	if !ok {
		return nil, fmt.Errorf("invalid type for Trojan")
	}
	p := &Proxy{
		Name:           data.Name,
		Type:           "trojan",
		Server:         data.Server,
		Port:           data.Port,
		Password:       password,
		SNI:            trojan.SecurityConfig.SNI,
		SkipCertVerify: trojan.SecurityConfig.AllowInsecure,
	}
	setTransportOptions(p, trojan.Transport, trojan.TransportConfig)
	return p, nil
}

func parseVless(data proxy.Proxy, uuid string) (*Proxy, error) {
	vless, ok := data.Option.(proxy.Vless)
	if !ok {
		return nil, fmt.Errorf("invalid type for Vless")
	}
	p := &Proxy{
		Name:   data.Name,
		Type:   "vless",
		Server: data.Server,
		Port:   data.Port,
		UUID:   uuid,
		Flow:   vless.Flow,
	}
	setSecurityOptions(p, vless.Security, vless.SecurityConfig)
	clashTransport(p, vless.Transport, vless.TransportConfig)
	return p, nil
}

func parseVmess(data proxy.Proxy, uuid string) (*Proxy, error) {
	vmess, ok := data.Option.(proxy.Vmess)
	if !ok {
		return nil, fmt.Errorf("invalid type for Vmess")
	}
	alterID := 0
	p := &Proxy{
		Name:    data.Name,
		Type:    "vmess",
		Server:  data.Server,
		Port:    data.Port,
		UUID:    uuid,
		AlterID: &alterID,
		Cipher:  "auto",
	}
	setSecurityOptions(p, vmess.Security, vmess.SecurityConfig)
	clashTransport(p, vmess.Transport, vmess.TransportConfig)
	return p, nil
}

func parseHysteria2(data proxy.Proxy, uuid string) (*Proxy, error) {
	hysteria2, ok := data.Option.(proxy.Hysteria2)
	if !ok {
		return nil, fmt.Errorf("invalid type for Hysteria2")
	}
	p := &Proxy{
		Name:              data.Name,
		Type:              "hysteria2",
		Server:            data.Server,
		Port:              data.Port,
		Ports:             hysteria2.HopPorts,
		Password:          uuid,
		HeartbeatInterval: hysteria2.HopInterval,
		SkipCertVerify:    hysteria2.SecurityConfig.AllowInsecure,
		SNI:               hysteria2.SecurityConfig.SNI,
	}
	if hysteria2.ObfsPassword != "" {
		p.Obfs = "salamander"
		p.ObfsPassword = hysteria2.ObfsPassword
	}

	return p, nil
}

func parseTuic(data proxy.Proxy, uuid string) (*Proxy, error) {
	tuic, ok := data.Option.(proxy.Tuic)
	if !ok {
		return nil, fmt.Errorf("invalid type for Tuic")
	}
	p := &Proxy{
		Name:           data.Name,
		Type:           "tuic",
		Server:         data.Server,
		Port:           data.Port,
		UUID:           uuid,
		Password:       uuid,
		SNI:            tuic.SecurityConfig.SNI,
		SkipCertVerify: tuic.SecurityConfig.AllowInsecure,
	}

	return p, nil
}

func setSecurityOptions(p *Proxy, security string, config proxy.SecurityConfig) {
	switch security {
	case "tls":
		p.TLS = true
		p.ServerName = config.SNI
		p.ClientFingerprint = config.Fingerprint
		p.SkipCertVerify = config.AllowInsecure
	case "reality":
		p.TLS = true
		p.ServerName = config.SNI
		p.ClientFingerprint = config.Fingerprint
		p.RealityOpts = RealityOptions{
			PublicKey: config.RealityPublicKey,
			ShortID:   config.RealityShortId,
		}
		p.SkipCertVerify = config.AllowInsecure
	default:
		p.TLS = false
	}
}

func setTransportOptions(p *Proxy, transport string, config proxy.TransportConfig) {
	switch transport {
	case "websocket":
		p.Network = "ws"
		p.WSOpts = WSOptions{
			Path: config.Path,
			Headers: map[string]string{
				"Host": config.Host,
			},
		}
	case "grpc":
		p.Network = "grpc"
		p.GrpcOpts = GrpcOptions{
			GrpcServiceName: config.ServiceName,
		}
	default:
		p.Network = "tcp"
	}
}
