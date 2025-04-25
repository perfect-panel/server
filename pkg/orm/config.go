package orm

import (
	"github.com/go-sql-driver/mysql"
)

func ParseDSN(dsn string) *Config {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil
	}
	return &Config{
		Addr:          cfg.Addr,
		Dbname:        cfg.DBName,
		Username:      cfg.User,
		Password:      cfg.Passwd,
		Config:        "charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai",
		MaxIdleConns:  10,
		MaxOpenConns:  10,
		SlowThreshold: 1000,
	}
}
