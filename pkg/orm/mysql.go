package orm

import (
	"errors"
	"fmt"
	"time"

	"github.com/perfect-panel/server/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Config struct {
	Addr          string `yaml:"Addr"`
	Username      string `yaml:"Username"`
	Password      string `yaml:"Password"`
	Dbname        string `yaml:"Dbname"`
	Config        string `yaml:"Config" default:"charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"`
	MaxIdleConns  int    `yaml:"MaxIdleConns" default:"10"`
	MaxOpenConns  int    `yaml:"MaxOpenConns" default:"10"`
	SlowThreshold int64  `yaml:"SlowThreshold" default:"1000"`
}

type Mysql struct {
	Config Config
}

func (m *Mysql) Dsn() string {
	return m.Config.Username + ":" + m.Config.Password + "@tcp(" + m.Config.Addr + ")/" + m.Config.Dbname + "?" + m.Config.Config
}

func (m *Mysql) GetSlowThreshold() time.Duration {
	return time.Duration(m.Config.SlowThreshold) * time.Millisecond
}
func (m *Mysql) GetColorful() bool {
	return true
}

func ConnectMysql(m Mysql) (*gorm.DB, error) {
	if m.Config.Dbname == "" {
		return nil, errors.New("database name is empty")
	}
	mysqlCfg := mysql.Config{
		DSN: m.Dsn(),
	}
	db, err := gorm.Open(mysql.New(mysqlCfg), &gorm.Config{
		Logger: new(logger.GormLogger),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	} else {
		sqldb, _ := db.DB()
		sqldb.SetMaxIdleConns(m.Config.MaxIdleConns)
		sqldb.SetMaxOpenConns(m.Config.MaxOpenConns)
		return db, nil
	}
}

func Ping(dsn string) bool {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("connect mysql failed, err: %v\n", err.Error())
		return false
	}
	sqlDB, _ := db.DB()
	return sqlDB.Ping() == nil
}
