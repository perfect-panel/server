package surge

import (
	"testing"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

func TestBuildHysteria2(t *testing.T) {
	tests := []struct {
		name     string
		data     proxy.Proxy
		uuid     string
		expected string
	}{
		{
			name: "Valid Hysteria2 with HopPorts",
			data: proxy.Proxy{
				Name:   "test",
				Server: "server.com",
				Port:   443,
				Option: proxy.Hysteria2{
					HopPorts: "1000-2000",
					SecurityConfig: proxy.SecurityConfig{
						SNI:           "example.com",
						AllowInsecure: true,
					},
				},
			},
			uuid:     "test-uuid",
			expected: "test=hysteria2,server.com,1000,password=test-uuid,udp-relay=true,sni=example.com,skip-cert-verify=true\r\n",
		},
		{
			name: "Valid Hysteria2 without HopPorts",
			data: proxy.Proxy{
				Name:   "test",
				Server: "server.com",
				Port:   443,
				Option: proxy.Hysteria2{
					SecurityConfig: proxy.SecurityConfig{
						SNI:           "example.com",
						AllowInsecure: false,
					},
				},
			},
			uuid:     "test-uuid",
			expected: "test=hysteria2,server.com,443,password=test-uuid,udp-relay=true,sni=example.com,skip-cert-verify=false\r\n",
		},
		{
			name: "Invalid Hysteria2 Option",
			data: proxy.Proxy{
				Name:   "test",
				Server: "server.com",
				Port:   443,
				Option: nil,
			},
			uuid:     "test-uuid",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildHysteria2(tt.data, tt.uuid)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
