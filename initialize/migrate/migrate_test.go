package migrate

import (
	"testing"

	"github.com/perfect-panel/ppanel-server/pkg/orm"
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
