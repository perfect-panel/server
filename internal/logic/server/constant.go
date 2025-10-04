package server

const (
	Unchanged   = "Unchanged"
	ShadowSocks = "shadowsocks"
	Vmess       = "vmess"
	Vless       = "vless"
	Trojan      = "trojan"
	AnyTLS      = "anytls"
	Tuic        = "tuic"
	Hysteria    = "hysteria"
	// Deprecated: Hysteria2 is deprecated, use Hysteria instead
	// TODO: remove in future versions
	Hysteria2 = "hysteria2"
)

type SecurityConfig struct {
	SNI                  string `json:"sni"`
	AllowInsecure        *bool  `json:"allow_insecure"`
	Fingerprint          string `json:"fingerprint"`
	RealityServerAddress string `json:"reality_server_addr"`
	RealityServerPort    int    `json:"reality_server_port"`
	RealityPrivateKey    string `json:"reality_private_key"`
	RealityPublicKey     string `json:"reality_public_key"`
	RealityShortId       string `json:"reality_short_id"`
	RealityMldsa65seed   string `json:"reality_mldsa65seed"`
}

type TransportConfig struct {
	Path                 string `json:"path"`
	Host                 string `json:"host"`
	ServiceName          string `json:"service_name"`
	DisableSNI           bool   `json:"disable_sni"`
	ReduceRtt            bool   `json:"reduce_rtt"`
	UDPRelayMode         string `json:"udp_relay_mode"`
	CongestionController string `json:"congestion_controller"`
}

type VlessNode struct {
	Port            uint16           `json:"port"`
	Flow            string           `json:"flow"`
	Network         string           `json:"transport"`
	TransportConfig *TransportConfig `json:"transport_config"`
	Security        string           `json:"security"`
	SecurityConfig  *SecurityConfig  `json:"security_config"`
}

type VmessNode struct {
	Port            uint16           `json:"port"`
	Network         string           `json:"transport"`
	TransportConfig *TransportConfig `json:"transport_config"`
	Security        string           `json:"security"`
	SecurityConfig  *SecurityConfig  `json:"security_config"`
}

type ShadowsocksNode struct {
	Port      uint16 `json:"port"`
	Cipher    string `json:"method"`
	ServerKey string `json:"server_key"`
}

type TrojanNode struct {
	Port            uint16           `json:"port"`
	Network         string           `json:"transport"`
	TransportConfig *TransportConfig `json:"transport_config"`
	Security        string           `json:"security"`
	SecurityConfig  *SecurityConfig  `json:"security_config"`
}

type AnyTLSNode struct {
	Port           uint16          `json:"port"`
	SecurityConfig *SecurityConfig `json:"security_config"`
}

type TuicNode struct {
	Port           uint16          `json:"port"`
	SecurityConfig *SecurityConfig `json:"security_config"`
}

type Hysteria2Node struct {
	Port           uint16          `json:"port"`
	HopPorts       string          `json:"hop_ports"`
	HopInterval    int             `json:"hop_interval"`
	ObfsPassword   string          `json:"obfs_password"`
	SecurityConfig *SecurityConfig `json:"security_config"`
}
