package user

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (m *defaultUserModel) UpdateUserSubscribeCache(ctx context.Context, data *Subscribe) error {
	return m.ClearSubscribeCacheByModels(ctx, data)
}

// QueryActiveSubscriptions returns the number of active subscriptions.
func (m *defaultUserModel) QueryActiveSubscriptions(ctx context.Context, subscribeId ...int64) (map[int64]int64, error) {
	type SubscriptionCount struct {
		SubscribeId int64
		Total       int64
	}
	var result []SubscriptionCount
	err := m.QueryNoCacheCtx(ctx, &result, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).
			Where("subscribe_id IN ? AND `status` IN ?", subscribeId, []int64{1, 0}).
			Select("subscribe_id, COUNT(id) as total").
			Group("subscribe_id").
			Scan(&result).
			Error
	})

	if err != nil {
		return nil, err
	}

	resultMap := make(map[int64]int64)
	for _, item := range result {
		resultMap[item.SubscribeId] = item.Total
	}

	return resultMap, nil
}

func (m *defaultUserModel) FindOneSubscribeByOrderId(ctx context.Context, orderId int64) (*Subscribe, error) {
	var data Subscribe
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("order_id = ?", orderId).First(&data).Error
	})
	return &data, err
}

func (m *defaultUserModel) FindOneSubscribe(ctx context.Context, id int64) (*Subscribe, error) {
	var data Subscribe
	key := fmt.Sprintf("%s%d", cacheUserSubscribeIdPrefix, id)
	err := m.QueryCtx(ctx, &data, key, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("id = ?", id).First(&data).Error
	})
	return &data, err
}

func (m *defaultUserModel) FindUsersSubscribeBySubscribeId(ctx context.Context, subscribeId int64) ([]*Subscribe, error) {
	var data []*Subscribe
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		err := conn.Model(&Subscribe{}).Where("subscribe_id = ? AND `status` IN ?", subscribeId, []int64{1, 0}).Find(v).Error

		if err != nil {
			return err
		}
		// update user subscribe status
		return conn.Model(&Subscribe{}).Where("subscribe_id = ? AND `status` = ?", subscribeId, 0).Update("status", 1).Error
	})
	return data, err
}

// QueryUserSubscribe returns a list of records that meet the conditions.
func (m *defaultUserModel) QueryUserSubscribe(ctx context.Context, userId int64, status ...int64) ([]*SubscribeDetails, error) {
	var list []*SubscribeDetails
	key := fmt.Sprintf("%s%d", cacheUserSubscribeUserPrefix, userId)
	err := m.QueryCtx(ctx, &list, key, func(conn *gorm.DB, v interface{}) error {
		// 获取当前时间
		now := time.Now()
		// 获取当前时间向前推 7 天
		sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
		// 基础条件查询
		conn = conn.Model(&Subscribe{}).Where("`user_id` = ?", userId)
		if len(status) > 0 {
			conn = conn.Where("`status` IN ?", status)
		}
		// 订阅过期时间大于当前时间或者订阅结束时间大于当前时间
		return conn.Where("`expire_time` > ? OR `finished_at` >= ? OR `expire_time` = ?", now, sevenDaysAgo, time.UnixMilli(0)).
			Preload("Subscribe").
			Find(&list).Error
	})
	return list, err
}

// FindOneUserSubscribe  finds a subscribeDetails by id.
func (m *defaultUserModel) FindOneUserSubscribe(ctx context.Context, id int64) (subscribeDetails *SubscribeDetails, err error) {
	//TODO cache
	//key := fmt.Sprintf("%s%d", cacheUserSubscribeUserPrefix, userId)
	err = m.QueryNoCacheCtx(ctx, subscribeDetails, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Preload("Subscribe").Where("id = ?", id).First(&subscribeDetails).Error
	})
	return
}

// FindOneSubscribeByToken  finds a record by token.
func (m *defaultUserModel) FindOneSubscribeByToken(ctx context.Context, token string) (*Subscribe, error) {
	var data Subscribe
	key := fmt.Sprintf("%s%s", cacheUserSubscribeTokenPrefix, token)
	err := m.QueryCtx(ctx, &data, key, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Where("token = ?", token).First(&data).Error
	})
	return &data, err
}

// UpdateSubscribe updates a record.
func (m *defaultUserModel) UpdateSubscribe(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error {
	old, err := m.FindOneSubscribe(ctx, data.Id)
	if err != nil {
		return err
	}

	// 使用 defer 确保更新后清理缓存
	defer func() {
		if clearErr := m.ClearSubscribeCacheByModels(ctx, old, data); clearErr != nil {
			// 记录清理缓存错误
		}
	}()

	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Model(&Subscribe{}).Where("id = ?", data.Id).Save(data).Error
	})
}

// DeleteSubscribe deletes a record.
func (m *defaultUserModel) DeleteSubscribe(ctx context.Context, token string, tx ...*gorm.DB) error {
	data, err := m.FindOneSubscribeByToken(ctx, token)
	if err != nil {
		return err
	}

	// 使用 defer 确保删除后清理缓存
	defer func() {
		if clearErr := m.ClearSubscribeCacheByModels(ctx, data); clearErr != nil {
			// 记录清理缓存错误
		}
	}()

	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Where("token = ?", token).Delete(&Subscribe{}).Error
	})
}

// InsertSubscribe insert Subscribe into the database.
func (m *defaultUserModel) InsertSubscribe(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error {
	// 使用 defer 确保插入后清理相关缓存
	defer func() {
		if clearErr := m.ClearSubscribeCacheByModels(ctx, data); clearErr != nil {
			// 记录清理缓存错误
		}
	}()

	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(data).Error
	})
}

func (m *defaultUserModel) DeleteSubscribeById(ctx context.Context, id int64, tx ...*gorm.DB) error {
	data, err := m.FindOneSubscribe(ctx, id)
	if err != nil {
		return err
	}

	// 使用 defer 确保删除后清理缓存
	defer func() {
		if clearErr := m.ClearSubscribeCacheByModels(ctx, data); clearErr != nil {
			// 记录清理缓存错误
		}
	}()

	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Where("id = ?", id).Delete(&Subscribe{}).Error
	})
}

func (m *defaultUserModel) ClearSubscribeCache(ctx context.Context, data ...*Subscribe) error {
	return m.ClearSubscribeCacheByModels(ctx, data...)
}
