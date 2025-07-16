package proxy

// Adapter represents a proxy adapter
type Adapter struct {
	Proxies []Proxy
	Group   []Group
	Rules   []string // rule
	Nodes   []string // all node
	Default string   // Default Node
}

// Proxy represents a proxy server
type Proxy struct {
	Name     string
	Server   string
	Port     int
	Protocol string
	Country  string
	Option   any
}

// Group represents a group of proxies
type Group struct {
	Name     string
	Type     GroupType
	Proxies  []string
	URL      string
	Interval int
	Direct   bool
}

type GroupType string

const (
	GroupTypeSelect   GroupType = "select"
	GroupTypeURLTest  GroupType = "url-test"
	GroupTypeFallback GroupType = "fallback"
)

// Shadowsocks represents a Shadowsocks proxy configuration
type Shadowsocks struct {
	Port      int    `json:"port"`
	Method    string `json:"method"`
	ServerKey string `json:"server_key"`
}

// Vless represents a Vless proxy configuration
type Vless struct {
	Port            int             `json:"port"`
	Flow            string          `json:"flow"`
	Transport       string          `json:"transport"`
	TransportConfig TransportConfig `json:"transport_config"`
	Security        string          `json:"security"`
	SecurityConfig  SecurityConfig  `json:"security_config"`
}

// Vmess represents a Vmess proxy configuration
type Vmess struct {
	Port            int             `json:"port"`
	Flow            string          `json:"flow"`
	Transport       string          `json:"transport"`
	TransportConfig TransportConfig `json:"transport_config"`
	Security        string          `json:"security"`
	SecurityConfig  SecurityConfig  `json:"security_config"`
}

// Trojan represents a Trojan proxy configuration
type Trojan struct {
	Port            int             `json:"port"`
	Flow            string          `json:"flow"`
	Transport       string          `json:"transport"`
	TransportConfig TransportConfig `json:"transport_config"`
	Security        string          `json:"security"`
	SecurityConfig  SecurityConfig  `json:"security_config"`
}

// Hysteria2 represents a Hysteria2 proxy configuration
type Hysteria2 struct {
	Port           int            `json:"port"`
	HopPorts       string         `json:"hop_ports"`
	HopInterval    int            `json:"hop_interval"`
	ObfsPassword   string         `json:"obfs_password"`
	SecurityConfig SecurityConfig `json:"security_config"`
}

// Tuic represents a Tuic proxy configuration
type Tuic struct {
	Port                 int            `json:"port"`
	DisableSNI           bool           `json:"disable_sni"`
	ReduceRtt            bool           `json:"reduce_rtt"`
	UDPRelayMode         string         `json:"udp_relay_mode"`
	CongestionController string         `json:"congestion_controller"`
	SecurityConfig       SecurityConfig `json:"security_config"`
}

// AnyTLS represents an AnyTLS proxy configuration
type AnyTLS struct {
	Port           int            `json:"port"`
	SecurityConfig SecurityConfig `json:"security_config"`
}

// TransportConfig represents the transport configuration for a proxy
type TransportConfig struct {
	Path                 string `json:"path,omitempty"` // ws/httpupgrade
	Host                 string `json:"host,omitempty"`
	ServiceName          string `json:"service_name"`          // grpc
	DisableSNI           bool   `json:"disable_sni"`           // Disable SNI for the transport(tuic)
	ReduceRtt            bool   `json:"reduce_rtt"`            // Reduce RTT for the transport(tuic)
	UDPRelayMode         string `json:"udp_relay_mode"`        // UDP relay mode for the transport(tuic)
	CongestionController string `json:"congestion_controller"` // Congestion controller for the transport(tuic)
}

// SecurityConfig represents the security configuration for a proxy
type SecurityConfig struct {
	SNI               string `json:"sni"`
	AllowInsecure     bool   `json:"allow_insecure"`
	Fingerprint       string `json:"fingerprint"`
	RealityServerAddr string `json:"reality_server_addr"`
	RealityServerPort int    `json:"reality_server_port"`
	RealityPrivateKey string `json:"reality_private_key"`
	RealityPublicKey  string `json:"reality_public_key"`
	RealityShortId    string `json:"reality_short_id"`
}

// Relay represents a relay configuration
type Relay struct {
	RelayHost    string
	DispatchMode string
	Prefix       string
}
