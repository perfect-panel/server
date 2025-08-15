package adapter

import (
	"testing"
	"time"
)

var tpl = `
{{- range $n := .Proxies }}
  {{- $dn := urlquery (default "node" $n.Name) -}}
  {{- $sni := default $n.Host $n.SNI -}}

  {{- if eq $n.Type "shadowsocks" -}}
    {{- $userinfo := b64enc (print $n.Method ":" $.UserInfo.Password) -}}
    {{- printf "ss://%s@%s:%v#%s" $userinfo $n.Host $n.Port $dn -}}
    {{- "\n" -}}
  {{- end -}}

  {{- if eq $n.Type "trojan" -}}
    {{- $qs := "security=tls" -}}
    {{- if $sni }}{{ $qs = printf "%s&sni=%s" $qs (urlquery $sni) }}{{ end -}}
    {{- if $n.AllowInsecure }}{{ $qs = printf "%s&allowInsecure=%v" $qs $n.AllowInsecure }}{{ end -}}
    {{- if $n.Fingerprint }}{{ $qs = printf "%s&fp=%s" $qs (urlquery $n.Fingerprint) }}{{ end -}}
    {{- printf "trojan://%s@%s:%v?%s#%s" $.UserInfo.Password $n.Host $n.Port $qs $dn -}}
    {{- "\n" -}}
  {{- end -}}

  {{- if eq $n.Type "vless" -}}
    {{- $qs := "encryption=none" -}}
    {{- if $n.RealityPublicKey -}}
      {{- $qs = printf "%s&security=reality" $qs -}}
      {{- $qs = printf "%s&pbk=%s" $qs (urlquery $n.RealityPublicKey) -}}
      {{- if $n.RealityShortId }}{{ $qs = printf "%s&sid=%s" $qs (urlquery $n.RealityShortId) }}{{ end -}}
    {{- else -}}
      {{- if or $n.SNI $n.Fingerprint $n.AllowInsecure }}
        {{- $qs = printf "%s&security=tls" $qs -}}
      {{- end -}}
    {{- end -}}
    {{- if $n.SNI }}{{ $qs = printf "%s&sni=%s" $qs (urlquery $n.SNI) }}{{ end -}}
    {{- if $n.AllowInsecure }}{{ $qs = printf "%s&allowInsecure=%v" $qs $n.AllowInsecure }}{{ end -}}
    {{- if $n.Fingerprint }}{{ $qs = printf "%s&fp=%s" $qs (urlquery $n.Fingerprint) }}{{ end -}}
    {{- if $n.Network }}{{ $qs = printf "%s&type=%s" $qs $n.Network }}{{ end -}}
    {{- if $n.Path }}{{ $qs = printf "%s&path=%s" $qs (urlquery $n.Path) }}{{ end -}}
    {{- if $n.ServiceName }}{{ $qs = printf "%s&serviceName=%s" $qs (urlquery $n.ServiceName) }}{{ end -}}
    {{- if $n.Flow }}{{ $qs = printf "%s&flow=%s" $qs (urlquery $n.Flow) }}{{ end -}}
    {{- printf "vless://%s@%s:%v?%s#%s" $n.ServerKey $n.Host $n.Port $qs $dn -}}
    {{- "\n" -}}
  {{- end -}}

  {{- if eq $n.Type "vmess" -}}
    {{- $obj := dict
        "v" "2"
        "ps" $n.Name
        "add" $n.Host
        "port" $n.Port
        "id" $n.ServerKey
        "aid" 0
        "net" (or $n.Network "tcp")
        "type" "none"
        "path" (or $n.Path "")
        "host" $n.Host
      -}}
    {{- if or $n.SNI $n.Fingerprint $n.AllowInsecure }}{{ set $obj "tls" "tls" }}{{ end -}}
    {{- if $n.SNI }}{{ set $obj "sni" $n.SNI }}{{ end -}}
    {{- if $n.Fingerprint }}{{ set $obj "fp" $n.Fingerprint }}{{ end -}}
    {{- printf "vmess://%s" (b64enc (toJson $obj)) -}}
    {{- "\n" -}}
  {{- end -}}

  {{- if or (eq $n.Type "hysteria2") (eq $n.Type "hy2") -}}
    {{- $qs := "" -}}
    {{- if $n.SNI }}{{ $qs = printf "sni=%s" (urlquery $n.SNI) }}{{ end -}}
    {{- if $n.AllowInsecure }}{{ $qs = printf "%s&insecure=%v" $qs $n.AllowInsecure }}{{ end -}}
    {{- if $n.ObfsPassword }}{{ $qs = printf "%s&obfs-password=%s" $qs (urlquery $n.ObfsPassword) }}{{ end -}}
    {{- printf "hy2://%s@%s:%v%s#%s"
          $.UserInfo.Password
          $n.Host
          $n.Port
          (ternary (gt (len $qs) 0) (print "?" $qs) "")
          $dn -}}
    {{- "\n" -}}
  {{- end -}}

  {{- if eq $n.Type "tuic" -}}
    {{- $qs := "" -}}
    {{- if $n.SNI }}{{ $qs = printf "sni=%s" (urlquery $n.SNI) }}{{ end -}}
    {{- if $n.AllowInsecure }}{{ $qs = printf "%s&insecure=%v" $qs $n.AllowInsecure }}{{ end -}}
    {{- printf "tuic://%s:%s@%s:%v%s#%s"
          $n.ServerKey
          $.UserInfo.Password
          $n.Host
          $n.Port
          (ternary (gt (len $qs) 0) (print "?" $qs) "")
          $dn -}}
    {{- "\n" -}}
  {{- end -}}

  {{- if eq $n.Type "anytls" -}}
    {{- $qs := "" -}}
    {{- if $n.SNI }}{{ $qs = printf "sni=%s" (urlquery $n.SNI) }}{{ end -}}
    {{- printf "anytls://%s@%s:%v%s#%s"
          $.UserInfo.Password
          $n.Host
          $n.Port
          (ternary (gt (len $qs) 0) (print "?" $qs) "")
          $dn -}}
    {{- "\n" -}}
  {{- end -}}

{{- end }}
`

func TestClient_Build(t *testing.T) {
	client := &Client{
		SiteName:       "TestSite",
		SubscribeName:  "TestSubscribe",
		ClientTemplate: tpl,
		Proxies: []Proxy{
			{
				Name:   "TestShadowSocks",
				Type:   "shadowsocks",
				Host:   "127.0.0.1",
				Port:   1234,
				Method: "aes-256-gcm",
			},
			{
				Name:          "TestTrojan",
				Type:          "trojan",
				Host:          "example.com",
				Port:          443,
				AllowInsecure: true,
				Security:      "tls",
				Transport:     "tcp",
				SNI:           "v1-dy.ixigua.com",
			},
		},
		UserInfo: User{
			Password:     "testpassword",
			ExpiredAt:    time.Now().Add(24 * time.Hour),
			Download:     1000000,
			Upload:       500000,
			Traffic:      1500000,
			SubscribeURL: "https://example.com/subscribe",
		},
	}
	buf, err := client.Build()
	if err != nil {
		t.Fatalf("Failed to build client: %v", err)
	}

	t.Logf("[测试] 输出: %s", buf)

}
