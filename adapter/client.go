package adapter

import (
	"bytes"
	"encoding/base64"
	"reflect"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
)

type Proxy struct {
	Name   string
	Server string
	Port   uint64
	Type   string
	Tags   []string

	// Security Options
	Security          string
	SNI               string // Server Name Indication for TLS
	AllowInsecure     bool   // Allow insecure connections (skip certificate verification)
	Fingerprint       string // Client fingerprint for TLS connections
	RealityServerAddr string // Reality server address
	RealityServerPort int    // Reality server port
	RealityPrivateKey string // Reality private key for authentication
	RealityPublicKey  string // Reality public key for authentication
	RealityShortId    string // Reality short ID for authentication
	// Transport Options
	Transport   string // Transport protocol (e.g., ws, http, grpc)
	Host        string // For WebSocket/HTTP/HTTPS
	Path        string // For HTTP/HTTPS
	ServiceName string // For gRPC
	// Shadowsocks Options
	Method    string
	ServerKey string // For Shadowsocks 2022
	// Vmess/Vless/Trojan Options
	Flow string // Flow for Vmess/Vless/Trojan
	// Hysteria2 Options
	HopPorts     string // Comma-separated list of hop ports
	HopInterval  int    // Interval for hop ports in seconds
	ObfsPassword string // Obfuscation password for Hysteria2

	// Tuic Options
	DisableSNI           bool   // Disable SNI
	ReduceRtt            bool   // Reduce RTT
	UDPRelayMode         string // UDP relay mode (e.g., "full", "partial")
	CongestionController string // Congestion controller (e.g., "cubic", "bbr")
}

type User struct {
	Password     string
	ExpiredAt    time.Time
	Download     int64
	Upload       int64
	Traffic      int64
	SubscribeURL string
}

type Client struct {
	SiteName       string  // Name of the site
	SubscribeName  string  // Name of the subscription
	ClientTemplate string  // Template for the entire client configuration
	OutputFormat   string  // json, yaml, etc.
	Proxies        []Proxy // List of proxy configurations
	UserInfo       User    // User information
}

func (c *Client) Build() ([]byte, error) {
	var buf bytes.Buffer
	tmpl, err := template.New("client").Funcs(sprig.TxtFuncMap()).Parse(c.ClientTemplate)
	if err != nil {
		return nil, err
	}

	proxies := make([]map[string]interface{}, len(c.Proxies))
	for i, p := range c.Proxies {
		proxies[i] = StructToMap(p)
	}

	err = tmpl.Execute(&buf, map[string]interface{}{
		"SiteName":      c.SiteName,
		"SubscribeName": c.SubscribeName,
		"OutputFormat":  c.OutputFormat,
		"Proxies":       proxies,
		"UserInfo":      c.UserInfo,
	})
	if err != nil {
		return nil, err
	}

	result := buf.String()
	if c.OutputFormat == "base64" {
		encoded := base64.StdEncoding.EncodeToString([]byte(result))
		return []byte(encoded), nil
	}

	return buf.Bytes(), nil
}

func StructToMap(obj interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		m[field.Name] = v.Field(i).Interface()
	}
	return m
}
