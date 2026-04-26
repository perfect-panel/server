package user

// V4.3 user_subscribe_device CRUD.
// Each user_subscribe owns N SubscribeDevice rows; tokens/uuids are per-device.

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/pkg/tool"
	"gorm.io/gorm"
)

const (
	cacheSubscribeDeviceIdPrefix     = "cache:subscribe_device:id:"
	cacheSubscribeDeviceTokenPrefix  = "cache:subscribe_device:token:"
	cacheSubscribeDeviceUuidPrefix   = "cache:subscribe_device:uuid:"
	cacheSubscribeDeviceListBySubKey = "cache:subscribe_device:by_sub:"
)

func (m *defaultUserModel) InsertSubscribeDevice(ctx context.Context, data *SubscribeDevice, tx ...*gorm.DB) error {
	if data.Token == "" {
		data.Token = tool.GenerateDeviceToken()
	}
	if data.UUID == "" {
		data.UUID = tool.GenerateUUIDv4()
	}
	if data.Status == 0 {
		data.Status = 1
	}
	defer func() {
		_ = m.DelCacheCtx(ctx, fmt.Sprintf("%s%d", cacheSubscribeDeviceListBySubKey, data.UserSubscribeId))
	}()
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(data).Error
	})
}

// BatchInsertSubscribeDevices — used by purchase/activation to create N slots in one statement.
func (m *defaultUserModel) BatchInsertSubscribeDevices(ctx context.Context, list []*SubscribeDevice, tx ...*gorm.DB) error {
	if len(list) == 0 {
		return nil
	}
	subId := list[0].UserSubscribeId
	for _, d := range list {
		if d.Token == "" {
			d.Token = tool.GenerateDeviceToken()
		}
		if d.UUID == "" {
			d.UUID = tool.GenerateUUIDv4()
		}
		if d.Status == 0 {
			d.Status = 1
		}
	}
	defer func() {
		_ = m.DelCacheCtx(ctx, fmt.Sprintf("%s%d", cacheSubscribeDeviceListBySubKey, subId))
	}()
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(&list).Error
	})
}

func (m *defaultUserModel) FindOneSubscribeDevice(ctx context.Context, id int64) (*SubscribeDevice, error) {
	var data SubscribeDevice
	key := fmt.Sprintf("%s%d", cacheSubscribeDeviceIdPrefix, id)
	err := m.QueryCtx(ctx, &data, key, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&SubscribeDevice{}).Where("id = ?", id).First(&data).Error
	})
	return &data, err
}

func (m *defaultUserModel) FindOneSubscribeDeviceByToken(ctx context.Context, token string) (*SubscribeDevice, error) {
	var data SubscribeDevice
	key := fmt.Sprintf("%s%s", cacheSubscribeDeviceTokenPrefix, token)
	err := m.QueryCtx(ctx, &data, key, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&SubscribeDevice{}).Where("token = ?", token).First(&data).Error
	})
	return &data, err
}

func (m *defaultUserModel) FindOneSubscribeDeviceByUUID(ctx context.Context, uuid string) (*SubscribeDevice, error) {
	var data SubscribeDevice
	key := fmt.Sprintf("%s%s", cacheSubscribeDeviceUuidPrefix, uuid)
	err := m.QueryCtx(ctx, &data, key, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&SubscribeDevice{}).Where("uuid = ?", uuid).First(&data).Error
	})
	return &data, err
}

func (m *defaultUserModel) QuerySubscribeDevices(ctx context.Context, userSubscribeId int64) ([]*SubscribeDevice, error) {
	var list []*SubscribeDevice
	key := fmt.Sprintf("%s%d", cacheSubscribeDeviceListBySubKey, userSubscribeId)
	err := m.QueryCtx(ctx, &list, key, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&SubscribeDevice{}).
			Where("user_subscribe_id = ?", userSubscribeId).
			Order("id ASC").
			Find(&list).Error
	})
	return list, err
}

func (m *defaultUserModel) UpdateSubscribeDevice(ctx context.Context, data *SubscribeDevice, tx ...*gorm.DB) error {
	old, err := m.FindOneSubscribeDevice(ctx, data.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	defer func() {
		keys := []string{
			fmt.Sprintf("%s%d", cacheSubscribeDeviceIdPrefix, data.Id),
			fmt.Sprintf("%s%d", cacheSubscribeDeviceListBySubKey, data.UserSubscribeId),
		}
		if old != nil {
			keys = append(keys,
				fmt.Sprintf("%s%s", cacheSubscribeDeviceTokenPrefix, old.Token),
				fmt.Sprintf("%s%s", cacheSubscribeDeviceUuidPrefix, old.UUID),
			)
		}
		keys = append(keys,
			fmt.Sprintf("%s%s", cacheSubscribeDeviceTokenPrefix, data.Token),
			fmt.Sprintf("%s%s", cacheSubscribeDeviceUuidPrefix, data.UUID),
		)
		_ = m.DelCacheCtx(ctx, keys...)
	}()
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Save(data).Error
	})
}

// DeleteSubscribeDevice — V4.3:删除单个设备槽。invoked by addon delete API.
// 调用方负责验权(归属 + is_addon)。这里只负责数据库 + 缓存清理。
func (m *defaultUserModel) DeleteSubscribeDevice(ctx context.Context, device *SubscribeDevice, tx ...*gorm.DB) error {
	defer func() {
		_ = m.DelCacheCtx(ctx,
			fmt.Sprintf("%s%d", cacheSubscribeDeviceIdPrefix, device.Id),
			fmt.Sprintf("%s%s", cacheSubscribeDeviceTokenPrefix, device.Token),
			fmt.Sprintf("%s%s", cacheSubscribeDeviceUuidPrefix, device.UUID),
			fmt.Sprintf("%s%d", cacheSubscribeDeviceListBySubKey, device.UserSubscribeId),
		)
	}()
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Where("id = ?", device.Id).Delete(&SubscribeDevice{}).Error
	})
}

// DeleteSubscribeDevicesBySubscribeId — invoked when a user_subscribe is deleted.
func (m *defaultUserModel) DeleteSubscribeDevicesBySubscribeId(ctx context.Context, userSubscribeId int64, tx ...*gorm.DB) error {
	defer func() {
		_ = m.DelCacheCtx(ctx, fmt.Sprintf("%s%d", cacheSubscribeDeviceListBySubKey, userSubscribeId))
	}()
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Where("user_subscribe_id = ?", userSubscribeId).Delete(&SubscribeDevice{}).Error
	})
}

// CountActiveDevicesBySubscribe — used to enforce max_device_count when user adds a slot.
func (m *defaultUserModel) CountActiveDevicesBySubscribe(ctx context.Context, userSubscribeId int64) (int64, error) {
	var count int64
	err := m.QueryNoCacheCtx(ctx, &count, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&SubscribeDevice{}).
			Where("user_subscribe_id = ?", userSubscribeId).
			Count(&count).Error
	})
	return count, err
}
