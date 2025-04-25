package orm

import "testing"

func TestParseDSN(t *testing.T) {
	dsn := "root:mylove520@tcp(localhost:3306)/vpnboard"
	config := ParseDSN(dsn)
	if config == nil {
		t.Fatal("config is nil")
	}
	t.Log(config)
}

func TestPing(t *testing.T) {
	dsn := "root:mylove520@tcp(localhost:3306)/vpnboard"
	status := Ping(dsn)
	t.Log(status)
}
