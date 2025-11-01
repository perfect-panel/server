package report

import (
	"testing"
)

func TestFreePort(t *testing.T) {
	port, err := FreePort()
	if err != nil {
		t.Fatalf("FreePort() error: %v", err)
	}
	t.Logf("FreePort: %v", port)
}

func TestModulePort(t *testing.T) {
	port, err := ModulePort()
	if err != nil {
		t.Fatalf("ModulePort() error: %v", err)
	}
	t.Logf("ModulePort: %v", port)
}
