package migrate

import (
	"testing"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/pkg/orm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func getDSN() string {

	cfg := orm.Config{
		Addr:     "127.0.0.1",
		Username: "root",
		Password: "mylove520",
		Dbname:   "vpnboard",
	}
	mc := orm.Mysql{
		Config: cfg,
	}
	return mc.Dsn()
}

func TestMigrate(t *testing.T) {
	t.Skipf("skip test")
	m := Migrate(getDSN())
	err := m.Migrate(2004)
	if err != nil {
		t.Errorf("failed to migrate: %v", err)
	} else {
		t.Log("migrate success")
	}
}
func TestMysql(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: "root:mylove520@tcp(localhost:3306)/vpnboard",
	}))
	if err != nil {
		t.Fatalf("Failed to connect to MySQL: %v", err)
	}
	err = db.Migrator().AutoMigrate(&node.Node{})
	if err != nil {
		t.Fatalf("Failed to auto migrate: %v", err)
		return
	}
	t.Log("MySQL connection and migration successful")
}
