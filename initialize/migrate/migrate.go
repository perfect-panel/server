package migrate

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/perfect-panel/server/pkg/logger"
)

//go:embed database/*.sql
var sqlFiles embed.FS
var NoChange = migrate.ErrNoChange

func Migrate(dsn string) *migrate.Migrate {
	d, err := iofs.New(sqlFiles, "database")
	if err != nil {
		logger.Errorf("[Migrate] iofs.New error: %v", err.Error())
		panic(err)
	}
	client, err := migrate.NewWithSourceInstance("iofs", d, fmt.Sprintf("mysql://%s", dsn))
	if err != nil {
		logger.Errorf("[Migrate] NewWithSourceInstance error: %v", err.Error())
		panic(err)
	}
	return client
}
