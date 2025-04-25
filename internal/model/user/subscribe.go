package user

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (m *defaultUserModel) UpdateUserSubscribeCache(ctx context.Context, data *Subscribe) error {
	return m.CachedConn.DelCacheCtx(ctx, m.getSubscribeCacheKey(data)...)
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
			Where("subscribe_id IN ? AND `status` IN ?", subscribeId, []int64{1, 0, 3}).
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
		return conn.Model(&Subscribe{}).Where("subscribe_id = ? AND `status` IN ?", subscribeId, []int64{1, 0}).Find(&data).Error
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
		conn = conn.Model(&Subscribe{}).Where("`user_id` = ? and `status` IN ?", userId, status)
		return conn.Where("`expire_time` > ? OR `finished_at` >= ?", now, sevenDaysAgo).
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
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Model(&Subscribe{}).Where("token = ?", data.Token).Save(data).Error
	}, m.getSubscribeCacheKey(data)...)
}

// DeleteSubscribe deletes a record.
func (m *defaultUserModel) DeleteSubscribe(ctx context.Context, token string, tx ...*gorm.DB) error {
	data, err := m.FindOneSubscribeByToken(ctx, token)
	if err != nil {
		return err
	}
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Where("token = ?", token).Delete(&Subscribe{}).Error
	}, m.getSubscribeCacheKey(data)...)
}

// InsertSubscribe insert Subscribe into the database.
func (m *defaultUserModel) InsertSubscribe(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error {
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(data).Error
	}, m.getSubscribeCacheKey(data)...)
}

func (m *defaultUserModel) DeleteSubscribeById(ctx context.Context, id int64, tx ...*gorm.DB) error {
	data, err := m.FindOneSubscribe(ctx, id)
	if err != nil {
		return err
	}
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Where("id = ?", id).Delete(&Subscribe{}).Error
	}, m.getSubscribeCacheKey(data)...)
}

func (m *defaultUserModel) ClearSubscribeCache(ctx context.Context, data ...*Subscribe) error {
	var keys []string
	for _, item := range data {
		keys = append(keys, m.getSubscribeCacheKey(item)...)
	}
	return m.CachedConn.DelCacheCtx(ctx, keys...)
}
