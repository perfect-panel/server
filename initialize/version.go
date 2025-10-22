package initialize

import (
	"errors"
	"time"

	"github.com/perfect-panel/server/internal/model/user"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/initialize/migrate"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/orm"
)

func Migrate(ctx *svc.ServiceContext) {
	mc := orm.Mysql{
		Config: ctx.Config.MySQL,
	}
	now := time.Now()
	if err := migrate.Migrate(mc.Dsn()).Up(); err != nil {
		if errors.Is(err, migrate.NoChange) {
			logger.Info("[Migrate] database not change")
			return
		}
		logger.Errorf("[Migrate] Up error: %v", err.Error())
		panic(err)
	} else {
		logger.Info("[Migrate] Database change, took " + time.Since(now).String())
	}
	// if not found admin user
	err := ctx.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&user.User{}).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := migrate.CreateAdminUser(ctx.Config.Administrator.Email, ctx.Config.Administrator.Password, tx); err != nil {
				logger.Errorf("[Migrate] CreateAdminUser error: %v", err.Error())
				return err
			}
			logger.Info("[Migrate] Create admin user success")
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
