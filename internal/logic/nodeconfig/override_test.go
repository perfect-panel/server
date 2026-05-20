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
