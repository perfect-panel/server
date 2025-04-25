package singbox

import (
	"github.com/perfect-panel/server/pkg/adapter/proxy"
)

type OutboundTLSOptions struct {
	Enabled         bool                    `json:"enabled,omitempty"`
	DisableSNI      bool                    `json:"disable_sni,omitempty"`
	ServerName      string                  `json:"server_name,omitempty"`
	Insecure        bool                    `json:"insecure,omitempty"`
	ALPN            Listable[string]        `json:"alpn,omitempty"`
	MinVersion      string                  `json:"min_version,omitempty"`
	MaxVersion      string                  `json:"max_version,omitempty"`
	CipherSuites    Listable[string]        `json:"cipher_suites,omitempty"`
	Certificate     Listable[string]        `json:"certificate,omitempty"`
	CertificatePath string                  `json:"certificate_path,omitempty"`
	ECH             *OutboundECHOptions     `json:"ech,omitempty"`
	UTLS            *OutboundUTLSOptions    `json:"utls,omitempty"`
	Reality         *OutboundRealityOptions `json:"reality,omitempty"`
}

func NewOutboundTLSOptions(security string, cfg proxy.SecurityConfig) *OutboundTLSOptions {
	var tls = &OutboundTLSOptions{}
	switch security {
	case "none":
		return nil
	case "tls":
		tls.Enabled = true
		if cfg.SNI != "" {
			tls.ServerName = cfg.SNI
		} else {
			tls.DisableSNI = true
		}
		tls.Insecure = cfg.AllowInsecure
		if cfg.Fingerprint != "" {
			tls.UTLS = &OutboundUTLSOptions{
				Enabled:     true,
				Fingerprint: cfg.Fingerprint,
			}
		}
	case "reality":
		tls.Enabled = true
		if cfg.SNI != "" {
			tls.ServerName = cfg.SNI
		} else {
			tls.DisableSNI = true
		}
		tls.Insecure = cfg.AllowInsecure
		if cfg.Fingerprint != "" {
			tls.UTLS = &OutboundUTLSOptions{
				Enabled:     true,
				Fingerprint: cfg.Fingerprint,
			}
		}
		tls.Reality = &OutboundRealityOptions{
			Enabled:   true,
			PublicKey: cfg.RealityPublicKey,
			ShortID:   cfg.RealityShortId,
		}
	}
	return tls
}

type OutboundECHOptions struct {
	Enabled                     bool             `json:"enabled,omitempty"`
	PQSignatureSchemesEnabled   bool             `json:"pq_signature_schemes_enabled,omitempty"`
	DynamicRecordSizingDisabled bool             `json:"dynamic_record_sizing_disabled,omitempty"`
	Config                      Listable[string] `json:"config,omitempty"`
	ConfigPath                  string           `json:"config_path,omitempty"`
}

type OutboundRealityOptions struct {
	Enabled   bool   `json:"enabled,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	ShortID   string `json:"short_id,omitempty"`
}

type OutboundUTLSOptions struct {
	Enabled     bool   `json:"enabled,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}
type Listable[T any] []T

type OutboundTLSOptionsContainer struct {
	TLS *OutboundTLSOptions `json:"tls,omitempty"`
}
