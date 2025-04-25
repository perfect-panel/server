package surge

import (
	"strings"
	"testing"
	"time"

	"github.com/perfect-panel/ppanel-server/pkg/adapter/proxy"
)

func TestSurgeBuild(t *testing.T) {
	adapter := proxy.Adapter{
		Proxies: []proxy.Proxy{
			{
				Name:     "test-shadowsocks",
				Protocol: "shadowsocks",
				Server:   "1.2.3.4",
				Port:     8388,
				Option: proxy.Shadowsocks{
					Method: "aes-256-gcm",
				},
			},
			{
				Name:     "test-trojan",
				Protocol: "trojan",
				Server:   "5.6.7.8",
				Port:     443,
				Option: proxy.Trojan{
					SecurityConfig: proxy.SecurityConfig{
						SNI:           "example.com",
						AllowInsecure: true,
					},
				},
			},
			{
				Name:     "test-hysteria",
				Protocol: "hysteria2",
				Server:   "1.1.1.1",
				Port:     443,
				Option: proxy.Hysteria2{
					HopPorts:    "8080-8090",
					HopInterval: 320,
					SecurityConfig: proxy.SecurityConfig{
						SNI:           "example.com",
						AllowInsecure: true,
					},
				},
			},
		},
		Group: []proxy.Group{
			{
				Name:    "test-group",
				Type:    proxy.GroupTypeSelect,
				Proxies: []string{"test-shadowsocks", "test-trojan", "test-hysteria"},
			},
			{
				Name:    "手动选择",
				Type:    proxy.GroupTypeSelect,
				Proxies: []string{"test-shadowsocks", "test-trojan", "test-hysteria"},
			},
		},
		Rules: []string{
			"DOMAIN-SUFFIX,example.com,DIRECT",
		},
	}

	user := UserInfo{
		UUID:         "test-uuid",
		Upload:       1024,
		Download:     2048,
		TotalTraffic: 4096,
		ExpiredDate:  time.Now().Add(24 * time.Hour),
		SubscribeURL: "http://example.com/subscribe",
	}

	surge := NewSurge(adapter)
	config := surge.Build("test-uuid", "TestSite", user)

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	configStr := string(config)
	t.Logf("configStr: %v", configStr)
	if !strings.Contains(configStr, "test-shadowsocks=ss") {
		t.Errorf("Expected config to contain test-shadowsocks proxy")
	}
	if !strings.Contains(configStr, "test-trojan=trojan") {
		t.Errorf("Expected config to contain test-trojan proxy")
	}
	if !strings.Contains(configStr, "test-group = select") {
		t.Errorf("Expected config to contain test-group proxy group")
	}
	if !strings.Contains(configStr, "DOMAIN-SUFFIX,example.com,DIRECT") {
		t.Errorf("Expected config to contain rule for example.com")
	}
}
