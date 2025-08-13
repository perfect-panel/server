package adapter

import (
	"testing"

	"github.com/perfect-panel/server/internal/model/server"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestAdapterProxy(t *testing.T) {

	servers := getServers()
	if len(servers) == 0 {
		t.Fatal("no servers found")
	}
	for _, srv := range servers {
		proxy, err := adapterProxy(*srv, "example.com", 0)
		if err != nil {
			t.Errorf("failed to adapt server %s: %v", srv.Name, err)
		}
		t.Logf("[测试] 适配服务器 %s 成功: %+v", srv.Name, proxy)
	}

}

func getServers() []*server.Server {
	db, err := connectMySQL("root:mylove520@tcp(localhost:3306)/perfectlink?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return nil
	}
	var servers []*server.Server
	if err = db.Model(&server.Server{}).Find(&servers).Error; err != nil {
		return nil
	}
	return servers
}
func connectMySQL(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
