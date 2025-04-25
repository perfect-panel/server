package general

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

type v2rayShareLink struct {
	Ps            string `json:"ps"`
	Add           string `json:"add"`
	Port          string `json:"port"`
	ID            string `json:"id"`
	Aid           string `json:"aid"`
	Net           string `json:"net"`
	Type          string `json:"type"`
	Host          string `json:"host"`
	SNI           string `json:"sni"`
	Path          string `json:"path"`
	TLS           string `json:"tls"`
	Flow          string `json:"flow,omitempty"`
	Alpn          string `json:"alpn,omitempty"`
	AllowInsecure bool   `json:"allowInsecure"`
	Fingerprint   string `json:"fp,omitempty"`
	PublicKey     string `json:"pbk,omitempty"`
	ShortId       string `json:"sid,omitempty"`
	SpiderX       string `json:"spx,omitempty"`
	V             string `json:"v"`
}

// GenerateBase64General  will output node URLs split by '\n' and then encode into base64
func GenerateBase64General(data []proxy.Proxy, uuid string) []byte {
	var links []string
	for _, v := range data {
		p := buildProxy(v, uuid)
		if p == "" {
			continue
		}
		links = append(links, p)
	}
	var rsp []byte
	rsp = base64.RawStdEncoding.AppendEncode(rsp, []byte(strings.Join(links, "\n")))
	return rsp
}

func buildProxy(data proxy.Proxy, uuid string) string {
	switch data.Protocol {
	case "shadowsocks":
		return ShadowsocksUri(data, uuid)
	case "vmess":
		return VmessUri(data, uuid)
	case "vless":
		return VlessUri(data, uuid)
	case "trojan":
		return TrojanUri(data, uuid)
	case "hysteria2":
		return Hysteria2Uri(data, uuid)
	case "tuic":
		return TuicUri(data, uuid)
	default:
		return ""
	}
}

func ShadowsocksUri(data proxy.Proxy, uuid string) string {
	ss := data.Option.(proxy.Shadowsocks)
	// sip002
	u := &url.URL{
		Scheme: "ss",
		// 还没有写 2022 的
		User:     url.User(strings.TrimSuffix(base64.URLEncoding.EncodeToString([]byte(ss.Method+":"+uuid)), "=")),
		Host:     net.JoinHostPort(data.Server, strconv.Itoa(data.Port)),
		Fragment: data.Name,
	}
	return u.String()
}

func VmessUri(data proxy.Proxy, uuid string) string {
	vmess := data.Option.(proxy.Vmess)

	transport := vmess.TransportConfig

	securityConfig := vmess.SecurityConfig

	var s = v2rayShareLink{
		V:    "2",
		Add:  data.Server,
		Port: fmt.Sprint(data.Port),
		ID:   uuid,
		Aid:  "0",
		Net:  vmess.Transport,
		// Type: "?",
		Host: transport.Host,
		Path: transport.Path,
	}

	if vmess.Security == "tls" {
		s.TLS = "tls"
		s.SNI = securityConfig.SNI
		s.AllowInsecure = securityConfig.AllowInsecure
		s.Fingerprint = securityConfig.Fingerprint
	}
	b, _ := json.Marshal(s)
	return "vmess://" + strings.TrimSuffix(base64.StdEncoding.EncodeToString(b), "=")
}

func VlessUri(data proxy.Proxy, uuid string) string {
	vless := data.Option.(proxy.Vless)
	transportConfig := vless.TransportConfig
	securityConfig := vless.SecurityConfig

	var query = make(url.Values)
	setQuery(&query, "flow", vless.Flow)
	setQuery(&query, "type", vless.Transport)
	setQuery(&query, "security", vless.Security)

	switch vless.Transport {
	case "ws", "http", "httpupgrade":
		setQuery(&query, "path", transportConfig.Path)
		setQuery(&query, "host", transportConfig.Host)
	case "grpc":
		setQuery(&query, "serviceName", transportConfig.ServiceName)
	case "meek":
		setQuery(&query, "url", transportConfig.Host)
	}

	setQuery(&query, "sni", securityConfig.SNI)
	setQuery(&query, "fp", securityConfig.Fingerprint)
	setQuery(&query, "pbk", securityConfig.RealityPublicKey)
	setQuery(&query, "sid", securityConfig.RealityShortId)

	u := url.URL{
		Scheme:   "vless",
		User:     url.User(uuid),
		Host:     net.JoinHostPort(data.Server, fmt.Sprint(data.Port)),
		RawQuery: query.Encode(),
		Fragment: data.Name,
	}
	return u.String()
}

func TrojanUri(data proxy.Proxy, uuid string) string {
	trojan := data.Option.(proxy.Trojan)
	transportConfig := trojan.TransportConfig
	securityConfig := trojan.SecurityConfig

	var query = make(url.Values)
	setQuery(&query, "type", trojan.Transport)
	setQuery(&query, "security", trojan.Security)

	switch trojan.Transport {
	case "ws", "http", "httpupgrade":
		setQuery(&query, "path", transportConfig.Path)
		setQuery(&query, "host", transportConfig.Host)
	case "grpc":
		setQuery(&query, "serviceName", transportConfig.ServiceName)
	case "meek":
		setQuery(&query, "url", transportConfig.Host)
	}

	setQuery(&query, "sni", securityConfig.SNI)
	setQuery(&query, "fp", securityConfig.Fingerprint)
	setQuery(&query, "pbk", securityConfig.RealityPublicKey)
	setQuery(&query, "sid", securityConfig.RealityShortId)

	if securityConfig.AllowInsecure {
		setQuery(&query, "allowInsecure", "1")
	}

	u := &url.URL{
		Scheme:   "trojan",
		User:     url.User(uuid),
		Host:     net.JoinHostPort(data.Server, strconv.Itoa(data.Port)),
		RawQuery: query.Encode(),
		Fragment: data.Name,
	}
	return u.String()
}

func Hysteria2Uri(data proxy.Proxy, uuid string) string {
	hysteria2 := data.Option.(proxy.Hysteria2)

	var query = make(url.Values)

	setQuery(&query, "sni", hysteria2.SecurityConfig.SNI)

	if hysteria2.SecurityConfig.AllowInsecure {
		setQuery(&query, "insecure", "1")
	}

	if hp := strings.TrimSpace(hysteria2.HopPorts); hp != "" {
		setQuery(&query, "mport", hp)
	}

	if hysteria2.ObfsPassword != "" {
		setQuery(&query, "obfs", "salamander")
		setQuery(&query, "obfs-password", hysteria2.ObfsPassword)
	}

	u := &url.URL{
		Scheme:   "hysteria2",
		User:     url.User(uuid),
		Host:     net.JoinHostPort(data.Server, strconv.Itoa(data.Port)),
		RawQuery: query.Encode(),
		Fragment: data.Name,
	}
	return u.String()
}

func TuicUri(data proxy.Proxy, uuid string) string {
	tuic := data.Option.(proxy.Tuic)
	var query = make(url.Values)

	setQuery(&query, "congestion_control", "bbr")

	if tuic.SecurityConfig.SNI == "" {
		setQuery(&query, "sni", tuic.SecurityConfig.SNI)
	} else {
		setQuery(&query, "disable_sni", "1")
	}
	if tuic.SecurityConfig.AllowInsecure {
		setQuery(&query, "allow_insecure", "1")
	}

	u := &url.URL{
		Scheme:   "tuic",
		User:     url.User(uuid + ":" + uuid),
		Host:     net.JoinHostPort(data.Server, strconv.Itoa(data.Port)),
		RawQuery: query.Encode(),
		Fragment: data.Name,
	}
	return u.String()
}

func setQuery(q *url.Values, k, v string) {
	if v != "" {
		q.Set(k, v)
	}
}
