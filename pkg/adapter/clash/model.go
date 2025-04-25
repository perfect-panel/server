package clash

type RawConfig struct {
	Port               int          `yaml:"port" json:"port"`
	SocksPort          int          `yaml:"socks-port" json:"socks-port"`
	RedirPort          int          `yaml:"redir-port" json:"redir-port"`
	TProxyPort         int          `yaml:"tproxy-port" json:"tproxy-port"`
	MixedPort          int          `yaml:"mixed-port" json:"mixed-port"`
	AllowLan           bool         `yaml:"allow-lan" json:"allow-lan"`
	Mode               string       `yaml:"mode" json:"mode"`
	LogLevel           string       `yaml:"log-level" json:"log-level"`
	ExternalController string       `yaml:"external-controller" json:"external-controller"`
	Secret             string       `yaml:"secret" json:"secret"`
	Proxies            []Proxy      `yaml:"proxies" json:"proxies"`
	ProxyGroups        []ProxyGroup `yaml:"proxy-groups" json:"proxy-groups"`
	Rules              []string     `yaml:"rules" json:"rule"`
}

type Proxy struct {
	// 基础数据
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Server string `yaml:"server"`
	Port   int    `yaml:"port,omitempty"`
	// Shadowsocks
	Password          string         `yaml:"password,omitempty"`
	Cipher            string         `yaml:"cipher,omitempty"`
	UDP               bool           `yaml:"udp,omitempty"`
	Plugin            string         `yaml:"plugin,omitempty"`
	PluginOpts        map[string]any `yaml:"plugin-opts,omitempty"`
	UDPOverTCP        bool           `yaml:"udp-over-tcp,omitempty"`
	UDPOverTCPVersion int            `yaml:"udp-over-tcp-version,omitempty"`
	ClientFingerprint string         `yaml:"client-fingerprint,omitempty"`
	// Vmess
	UUID                string         `yaml:"uuid,omitempty"`
	AlterID             *int           `yaml:"alterId,omitempty"`
	Network             string         `yaml:"network,omitempty"`
	TLS                 bool           `yaml:"tls,omitempty"`
	ALPN                []string       `yaml:"alpn,omitempty"`
	SkipCertVerify      bool           `yaml:"skip-cert-verify,omitempty"`
	Fingerprint         string         `yaml:"fingerprint,omitempty"`
	ServerName          string         `yaml:"servername,omitempty"`
	RealityOpts         RealityOptions `yaml:"reality-opts,omitempty"`
	HTTPOpts            HTTPOptions    `yaml:"http-opts,omitempty"`
	HTTP2Opts           HTTP2Options   `yaml:"h2-opts,omitempty"`
	GrpcOpts            GrpcOptions    `yaml:"grpc-opts,omitempty"`
	WSOpts              WSOptions      `yaml:"ws-opts,omitempty"`
	PacketAddr          bool           `yaml:"packet-addr,omitempty"`
	XUDP                bool           `yaml:"xudp,omitempty"`
	PacketEncoding      string         `yaml:"packet-encoding,omitempty"`
	GlobalPadding       bool           `yaml:"global-padding,omitempty"`
	AuthenticatedLength bool           `yaml:"authenticated-length,omitempty"`
	// Vless
	Flow      string            `yaml:"flow,omitempty"`
	WSPath    string            `yaml:"ws-path,omitempty"`
	WSHeaders map[string]string `yaml:"ws-headers,omitempty"`
	// Trojan
	SNI    string         `yaml:"sni,omitempty"`
	SSOpts TrojanSSOption `yaml:"ss-opts,omitempty"`
	// Hysteria2
	Ports          string `yaml:"ports,omitempty"`
	HopInterval    int    `yaml:"hop-interval,omitempty"`
	Up             string `yaml:"up,omitempty"`
	Down           string `yaml:"down,omitempty"`
	Obfs           string `yaml:"obfs,omitempty"`
	ObfsPassword   string `yaml:"obfs-password,omitempty"`
	CustomCA       string `yaml:"ca,omitempty"`
	CustomCAString string `yaml:"ca-str,omitempty"`
	CWND           int    `yaml:"cwnd,omitempty"`
	UdpMTU         int    `yaml:"udp-mtu,omitempty"`
	// Tuic
	Token                 string `yaml:"token,omitempty"`
	Ip                    string `yaml:"ip,omitempty"`
	HeartbeatInterval     int    `yaml:"heartbeat-interval,omitempty"`
	ReduceRtt             bool   `yaml:"reduce-rtt,omitempty"`
	RequestTimeout        int    `yaml:"request-timeout,omitempty"`
	UdpRelayMode          string `yaml:"udp-relay-mode,omitempty"`
	CongestionController  string `yaml:"congestion-controller,omitempty"`
	DisableSni            bool   `yaml:"disable-sni,omitempty"`
	MaxUdpRelayPacketSize int    `yaml:"max-udp-relay-packet-size,omitempty"`
	FastOpen              bool   `yaml:"fast-open,omitempty"`
	MaxOpenStreams        int    `yaml:"max-open-streams,omitempty"`
	ReceiveWindowConn     int    `yaml:"recv-window-conn,omitempty"`
	ReceiveWindow         int    `yaml:"recv-window,omitempty"`
	DisableMTUDiscovery   bool   `yaml:"disable-mtu-discovery,omitempty"`
	MaxDatagramFrameSize  int    `yaml:"max-datagram-frame-size,omitempty"`
	UDPOverStream         bool   `yaml:"udp-over-stream,omitempty"`
	UDPOverStreamVersion  int    `yaml:"udp-over-stream-version,omitempty"`
}
type ProxyGroup struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Proxies  []string `yaml:"proxies"`
	Url      string   `yaml:"url,omitempty"`
	Interval int      `yaml:"interval,omitempty"`
}

type TrojanSSOption struct {
	Enabled  bool   `yaml:"enabled,omitempty"`
	Method   string `yaml:"method,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type RealityOptions struct {
	PublicKey string `yaml:"public-key"`
	ShortID   string `yaml:"short-id"`
}

type HTTPOptions struct {
	Method  string              `yaml:"method,omitempty"`
	Path    []string            `yaml:"path,omitempty"`
	Headers map[string][]string `yaml:"headers,omitempty"`
}

type HTTP2Options struct {
	Host []string `yaml:"host,omitempty"`
	Path string   `yaml:"path,omitempty"`
}

type GrpcOptions struct {
	GrpcServiceName string `yaml:"grpc-service-name,omitempty"`
}

type WSOptions struct {
	Path                     string            `yaml:"path,omitempty"`
	Headers                  map[string]string `yaml:"headers,omitempty"`
	MaxEarlyData             int               `yaml:"max-early-data,omitempty"`
	EarlyDataHeaderName      string            `yaml:"early-data-header-name,omitempty"`
	V2rayHttpUpgrade         bool              `yaml:"v2ray-http-upgrade,omitempty"`
	V2rayHttpUpgradeFastOpen bool              `yaml:"v2ray-http-upgrade-fast-open,omitempty"`
}
