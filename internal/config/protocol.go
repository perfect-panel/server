package config

type Protocol string

const (
	Shadowsocks Protocol = "shadowsocks"
	Trojan      Protocol = "trojan"
	Vmess       Protocol = "vmess"
	Vless       Protocol = "vless"
)
