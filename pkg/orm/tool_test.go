package orm

import (
	"os"
	"strings"
	"testing"
)

func TestParseMySQLDSN(t *testing.T) {
	cfg := ParseDSN("root:password@tcp(localhost:3306)/ppanel?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai")
	if cfg == nil {
		t.Fatal("config is nil")
	}
	if cfg.Driver != DriverMySQL || cfg.Addr != "localhost:3306" || cfg.Dbname != "ppanel" || cfg.Username != "root" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestParsePostgresDSN(t *testing.T) {
	cfg := ParseDSN("postgres://postgres:password@localhost:5432/ppanel?sslmode=disable&TimeZone=Asia%2FShanghai")
	if cfg == nil {
		t.Fatal("config is nil")
	}
	if cfg.Driver != DriverPostgres || cfg.Addr != "localhost:5432" || cfg.Dbname != "ppanel" || cfg.Username != "postgres" {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	dsn := Mysql{Config: *cfg}.Dsn()
	if want := "TimeZone=Asia/Shanghai"; !strings.Contains(dsn, want) {
		t.Fatalf("postgres dsn %q does not contain %q", dsn, want)
	}
}

func TestPingMySQL(t *testing.T) {
	dsn := os.Getenv("PPANEL_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set PPANEL_TEST_MYSQL_DSN to run MySQL/MariaDB ping test")
	}
	if !PingDatabase(DriverMySQL, dsn) {
		t.Fatal("mysql ping failed")
	}
}

func TestPingPostgres(t *testing.T) {
	dsn := os.Getenv("PPANEL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set PPANEL_TEST_POSTGRES_DSN to run PostgreSQL ping test")
	}
	if !PingDatabase(DriverPostgres, dsn) {
		t.Fatal("postgres ping failed")
	}
}
