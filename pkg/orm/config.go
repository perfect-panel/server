package orm

import (
	"net/url"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func ParseDSN(dsn string) *Config {
	if cfg := parseURLDSN(dsn); cfg != nil {
		return cfg
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil
	}
	return &Config{
		Driver:        DriverMySQL,
		Addr:          cfg.Addr,
		Dbname:        cfg.DBName,
		Username:      cfg.User,
		Password:      cfg.Passwd,
		Config:        DefaultMySQLConfig,
		MaxIdleConns:  10,
		MaxOpenConns:  10,
		SlowThreshold: DefaultSlowThresholdMs,
	}
}

func parseURLDSN(dsn string) *Config {
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		return nil
	}
	driver := NormalizeDriver(u.Scheme)
	if driver != DriverMySQL && driver != DriverPostgres {
		return nil
	}
	password, _ := u.User.Password()
	query := u.RawQuery
	if query == "" {
		if driver == DriverPostgres {
			query = DefaultPostgresConfig
		} else {
			query = DefaultMySQLConfig
		}
	}
	return &Config{
		Driver:        driver,
		Addr:          u.Host,
		Dbname:        strings.TrimPrefix(u.Path, "/"),
		Username:      u.User.Username(),
		Password:      password,
		Config:        query,
		MaxIdleConns:  10,
		MaxOpenConns:  10,
		SlowThreshold: DefaultSlowThresholdMs,
	}
}
