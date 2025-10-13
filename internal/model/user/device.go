package user

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func (m *customUserModel) FindOneDevice(ctx context.Context, id int64) (*Device, error) {
	deviceIdKey := fmt.Sprintf("%s%v", cacheUserDeviceIdPrefix, id)
	var resp Device
	err := m.QueryCtx(ctx, &resp, deviceIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Device{}).Where("`id` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *customUserModel) FindOneDeviceByIdentifier(ctx context.Context, id string) (*Device, error) {
	deviceIdKey := fmt.Sprintf("%s%v", cacheUserDeviceNumberPrefix, id)
	var resp Device
	err := m.QueryCtx(ctx, &resp, deviceIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Device{}).Where("`identifier` = ?", id).First(&resp).Error
	})
	switch {
	case err == nil:
		return &resp, nil
	default:
		return nil, err
	}
}

// QueryDevicePageList  returns a list of records that meet the conditions.
func (m *customUserModel) QueryDevicePageList(ctx context.Context, userId, subscribeId int64, page, size int) ([]*Device, int64, error) {
	var list []*Device
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Device{}).Where("`user_id` = ? and `subscribe_id` = ?", userId, subscribeId).Count(&total).Limit(size).Offset((page - 1) * size).Find(&list).Error
	})
	return list, total, err
}

// QueryDeviceList  returns a list of records that meet the conditions.
func (m *customUserModel) QueryDeviceList(ctx context.Context, userId int64) ([]*Device, int64, error) {
	var list []*Device
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Device{}).Where("`user_id` = ?", userId).Count(&total).Find(&list).Error
	})
	return list, total, err
}

func (m *customUserModel) UpdateDevice(ctx context.Context, data *Device, tx ...*gorm.DB) error {
	old, err := m.FindOneDevice(ctx, data.Id)
	if err != nil {
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Save(data).Error
	}, old.GetCacheKeys()...)
	return err
}

func (m *customUserModel) DeleteDevice(ctx context.Context, id int64, tx ...*gorm.DB) error {
	data, err := m.FindOneDevice(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Delete(&Device{}, id).Error
	}, data.GetCacheKeys()...)
	return err
}

func (m *customUserModel) InsertDevice(ctx context.Context, data *Device, tx ...*gorm.DB) error {
	defer func() {
		if clearErr := m.ClearDeviceCache(ctx, data); clearErr != nil {
			// log cache clear error
		}
	}()

	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(data).Error
	})
}
