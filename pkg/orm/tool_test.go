package orm

import (
	"testing"

	"github.com/perfect-panel/server/internal/model/task"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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

func TestMysql(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: "root:mylove520@tcp(localhost:3306)/vpnboard",
	}))
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}
	err = db.Migrator().AutoMigrate(&task.Task{})
	if err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
		return
	}
	t.Log("MySQL connection and migration successful")
}
