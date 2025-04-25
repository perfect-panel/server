package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/ppanel-server/internal/config"
	"github.com/perfect-panel/ppanel-server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var _ Model = (*customApplicationModel)(nil)
var (
	cacheApplicationIdPrefix        = "cache:application:id:"
	cacheApplicationConfigIdPrefix  = "cache:application:config:id:"
	cacheApplicationVersionIdPrefix = "cache:application:version:id:"
)

type (
	Model interface {
		applicationModel
		customApplicationLogicModel
	}
	applicationModel interface {
		Insert(ctx context.Context, data *Application) error
		FindOne(ctx context.Context, id int64) (*Application, error)
		Update(ctx context.Context, data *Application) error
		Delete(ctx context.Context, id int64) error
		InsertVersion(ctx context.Context, data *ApplicationVersion) error
		FindOneVersion(ctx context.Context, id int64) (*ApplicationVersion, error)
		UpdateVersion(ctx context.Context, data *ApplicationVersion) error
		InsertConfig(ctx context.Context, data *ApplicationConfig) error
		FindOneConfig(ctx context.Context, id int64) (*ApplicationConfig, error)
		UpdateConfig(ctx context.Context, data *ApplicationConfig) error
		DeleteVersion(ctx context.Context, id int64) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customApplicationModel struct {
		*defaultApplicationModel
	}
	defaultApplicationModel struct {
		cache.CachedConn
		table string
	}
)

func newApplicationModel(db *gorm.DB, c *redis.Client) *defaultApplicationModel {
	return &defaultApplicationModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`Application`",
	}
}

func (m *defaultApplicationModel) getCacheKeys(data *Application) []string {
	if data == nil {
		return []string{}
	}
	ApplicationIdKey := fmt.Sprintf("%s%v", cacheApplicationIdPrefix, data.Id)
	cacheKeys := []string{
		ApplicationIdKey,
		config.ApplicationKey,
	}
	return cacheKeys
}

func (m *defaultApplicationModel) Insert(ctx context.Context, data *Application) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultApplicationModel) FindOne(ctx context.Context, id int64) (*Application, error) {
	ApplicationIdKey := fmt.Sprintf("%s%v", cacheApplicationIdPrefix, id)
	var resp Application
	err := m.QueryCtx(ctx, &resp, ApplicationIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Application{}).Preload("ApplicationVersions").Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultApplicationModel) Update(ctx context.Context, data *Application) error {
	old, err := m.FindOne(ctx, data.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Save(data).Error
	}, m.getCacheKeys(old)...)
	return err
}

func (m *defaultApplicationModel) Delete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		err = db.Where("application_id = ?", id).Delete(&ApplicationVersion{}).Error
		if err != nil {
			return err
		}
		return db.Delete(&Application{}, id).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultApplicationModel) getVersionCacheKeys(data *ApplicationVersion) []string {
	if data == nil {
		return []string{}
	}
	ApplicationVersionIdKey := fmt.Sprintf("%s%v", cacheApplicationVersionIdPrefix, data.Id)
	cacheKeys := []string{
		ApplicationVersionIdKey,
		config.ApplicationKey,
	}
	return cacheKeys
}
func (m *defaultApplicationModel) getConfigCacheKeys(data *ApplicationConfig) []string {
	if data == nil {
		return []string{}
	}
	ApplicationConfigIdKey := fmt.Sprintf("%s%v", cacheApplicationConfigIdPrefix, data.Id)
	cacheKeys := []string{
		ApplicationConfigIdKey,
		config.ApplicationKey,
	}
	return cacheKeys
}

func (m *defaultApplicationModel) InsertVersion(ctx context.Context, data *ApplicationVersion) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Transaction(func(tx *gorm.DB) error {
			if data.IsDefault {
				err := tx.Model(&ApplicationVersion{}).
					Where("application_id = ? and platform = ? and default_version = ?", data.ApplicationId, data.Platform, data.IsDefault).
					Updates(map[string]interface{}{"default_version": false}).Error
				if err != nil {
					return err
				}
			}
			return tx.Create(&data).Error
		})
	}, m.getVersionCacheKeys(data)...)
	return err
}

func (m *defaultApplicationModel) FindOneVersion(ctx context.Context, id int64) (*ApplicationVersion, error) {
	ApplicationVersionIdKey := fmt.Sprintf("%s%v", cacheApplicationVersionIdPrefix, id)
	var resp ApplicationVersion
	err := m.QueryCtx(ctx, &resp, ApplicationVersionIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&ApplicationVersion{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultApplicationModel) UpdateVersion(ctx context.Context, data *ApplicationVersion) error {
	old, err := m.FindOneVersion(ctx, data.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Transaction(func(tx *gorm.DB) error {
			if data.IsDefault {
				err := tx.Model(&ApplicationVersion{}).
					Where("application_id = ? and platform = ? and default_version = ?", data.ApplicationId, data.Platform, data.IsDefault).
					Updates(map[string]interface{}{"default_version": false}).Error
				if err != nil {
					return err
				}
			}
			return tx.Save(data).Error
		})
	}, m.getVersionCacheKeys(old)...)
	return err
}

func (m *defaultApplicationModel) InsertConfig(ctx context.Context, data *ApplicationConfig) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Create(&data).Error
	}, m.getConfigCacheKeys(data)...)
	return err
}

func (m *defaultApplicationModel) FindOneConfig(ctx context.Context, id int64) (*ApplicationConfig, error) {
	ApplicationConfigIdKey := fmt.Sprintf("%s%v", cacheApplicationConfigIdPrefix, id)
	var resp ApplicationConfig
	err := m.QueryCtx(ctx, &resp, ApplicationConfigIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&ApplicationConfig{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *defaultApplicationModel) UpdateConfig(ctx context.Context, data *ApplicationConfig) error {
	old, err := m.FindOneConfig(ctx, data.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Save(data).Error
	}, m.getConfigCacheKeys(old)...)
	return err
}

func (m *defaultApplicationModel) DeleteVersion(ctx context.Context, id int64) error {
	data, err := m.FindOneVersion(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		db := conn
		return db.Delete(&ApplicationVersion{}, id).Error
	}, m.getVersionCacheKeys(data)...)
	return err
}

func (m *defaultApplicationModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
