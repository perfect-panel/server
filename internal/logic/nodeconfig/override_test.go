package nodeconfig

import (
	"testing"

	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/types"
)

func TestOverrideModelAndApplyOverride(t *testing.T) {
	global := GlobalValues(config.NodeConfig{
		IPStrategy: "prefer_ipv4",
		DNS: []config.NodeDNS{
			{Proto: "udp", Address: "8.8.8.8:53", Domains: []string{"geosite:google"}},
		},
		Block: []string{"geosite:ads"},
		Outbound: []config.NodeOutbound{
			{Name: "global", Protocol: "direct", Rules: []string{"geoip:cn"}},
		},
	})

	override, allInherited, err := OverrideModel(1, types.ServerNodeConfigOverride{
		InheritIPStrategy: false,
		IPStrategy:        "prefer_ipv6",
		InheritDNS:        false,
		DNS:               []types.NodeDNS{},
		InheritBlock:      true,
		InheritOutbound:   false,
		Outbound: []types.NodeOutbound{
			{Name: "node", Protocol: "reject", Rules: []string{"geosite:private"}},
		},
	})
	if err != nil {
		t.Fatalf("OverrideModel() error = %v", err)
	}
	if allInherited {
		t.Fatal("OverrideModel() allInherited = true, want false")
	}
	override.Id = 1

	effective := CloneValues(global)
	if err := ApplyOverride(&effective, override); err != nil {
		t.Fatalf("ApplyOverride() error = %v", err)
	}
	if effective.IPStrategy != "prefer_ipv6" {
		t.Fatalf("IPStrategy = %q, want prefer_ipv6", effective.IPStrategy)
	}
	if len(effective.DNS) != 0 {
		t.Fatalf("DNS len = %d, want 0", len(effective.DNS))
	}
	if len(effective.Block) != 1 || effective.Block[0] != "geosite:ads" {
		t.Fatalf("Block = %#v, want inherited global block", effective.Block)
	}
	if len(effective.Outbound) != 1 || effective.Outbound[0].Name != "node" {
		t.Fatalf("Outbound = %#v, want node override", effective.Outbound)
	}
}

func TestOverrideModelAllInherited(t *testing.T) {
	_, allInherited, err := OverrideModel(1, types.ServerNodeConfigOverride{
		InheritIPStrategy: true,
		InheritDNS:        true,
		InheritBlock:      true,
		InheritOutbound:   true,
	})
	if err != nil {
		t.Fatalf("OverrideModel() error = %v", err)
	}
	if !allInherited {
		t.Fatal("OverrideModel() allInherited = false, want true")
	}
}

func TestSanitizeNodeConfigValues(t *testing.T) {
	global := GlobalValues(config.NodeConfig{
		IPStrategy: "prefer_ipv4",
		DNS: []config.NodeDNS{
			{Proto: " udp ", Address: " 8.8.8.8:53 ", Domains: []string{" geosite:google ", "", "geosite:google"}},
			{Proto: "", Address: "1.1.1.1:53", Domains: []string{"geosite:cloudflare"}},
		},
		Block: []string{" geosite:ads ", "", "geosite:ads", "   "},
		Outbound: []config.NodeOutbound{
			{Name: " node ", Protocol: " socks ", Address: " 127.0.0.1 ", Rules: []string{" geoip:private ", "", "geoip:private"}},
			{Name: "empty-rules", Protocol: "direct", Rules: []string{" "}},
			{Name: "", Protocol: "direct", Rules: []string{"geoip:cn"}},
		},
	})

	if len(global.DNS) != 1 {
		t.Fatalf("DNS len = %d, want 1", len(global.DNS))
	}
	if global.DNS[0].Proto != "udp" || global.DNS[0].Address != "8.8.8.8:53" {
		t.Fatalf("DNS = %#v, want trimmed valid DNS", global.DNS[0])
	}
	if len(global.DNS[0].Domains) != 1 || global.DNS[0].Domains[0] != "geosite:google" {
		t.Fatalf("DNS domains = %#v, want sanitized domains", global.DNS[0].Domains)
	}
	if len(global.Block) != 1 || global.Block[0] != "geosite:ads" {
		t.Fatalf("Block = %#v, want sanitized block rules", global.Block)
	}
	if len(global.Outbound) != 1 {
		t.Fatalf("Outbound len = %d, want 1", len(global.Outbound))
	}
	if global.Outbound[0].Name != "node" || global.Outbound[0].Protocol != "socks" {
		t.Fatalf("Outbound = %#v, want trimmed outbound", global.Outbound[0])
	}
	if len(global.Outbound[0].Rules) != 1 || global.Outbound[0].Rules[0] != "geoip:private" {
		t.Fatalf("Outbound rules = %#v, want sanitized rules", global.Outbound[0].Rules)
	}
}
