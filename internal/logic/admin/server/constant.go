package server

const (
	ShadowSocks = "shadowsocks"
	Vmess       = "vmess"
	Vless       = "vless"
	Trojan      = "trojan"
	AnyTLS      = "anytls"
	Tuic        = "tuic"
	// Hysteria is the canonical key for the Hysteria 2 protocol.
	// The wire value stays "hysteria" for backward compatibility;
	// admin UI shows "Hysteria 2".
	Hysteria = "hysteria"
)
