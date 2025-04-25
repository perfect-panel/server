package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/ppanel-server/pkg/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	cacheUserIdPrefix    = "cache:user:id:"
	cacheUserEmailPrefix = "cache:user:email:"
)
var _ Model = (*customUserModel)(nil)

type (
	Model interface {
		userModel
		customUserLogicModel
	}
	userModel interface {
		Insert(ctx context.Context, data *User, tx ...*gorm.DB) error
		FindOne(ctx context.Context, id int64) (*User, error)
		Update(ctx context.Context, data *User, tx ...*gorm.DB) error
		Delete(ctx context.Context, id int64, tx ...*gorm.DB) error
		Transaction(ctx context.Context, fn func(db *gorm.DB) error) error
	}

	customUserModel struct {
		*defaultUserModel
	}
	defaultUserModel struct {
		cache.CachedConn
		table string
	}
)

func newUserModel(db *gorm.DB, c *redis.Client) *defaultUserModel {
	return &defaultUserModel{
		CachedConn: cache.NewConn(db, c),
		table:      "`user`",
	}
}

func (m *defaultUserModel) batchGetCacheKeys(users ...*User) []string {
	var keys []string
	for _, user := range users {
		keys = append(keys, m.getCacheKeys(user)...)
	}
	return keys

}
func (m *defaultUserModel) getCacheKeys(data *User) []string {
	if data == nil {
		return []string{}
	}
	userIdKey := fmt.Sprintf("%s%v", cacheUserIdPrefix, data.Id)
	cacheKeys := []string{
		userIdKey,
	}
	// email key
	if len(data.AuthMethods) > 0 {
		for _, auth := range data.AuthMethods {
			if auth.AuthType == "email" {
				cacheKeys = append(cacheKeys, fmt.Sprintf("%s%v", cacheUserEmailPrefix, auth.AuthIdentifier))
				break
			}
		}
	}
	return cacheKeys
}

func (m *defaultUserModel) FindOneByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	key := fmt.Sprintf("%s%v", cacheUserEmailPrefix, email)
	err := m.QueryCtx(ctx, &user, key, func(conn *gorm.DB, v interface{}) error {
		var data AuthMethods
		if err := conn.Model(&AuthMethods{}).Where("`auth_type` = 'email' AND `auth_identifier` = ?", email).First(&data).Error; err != nil {
			return err
		}
		return conn.Model(&User{}).Where("`id` = ?", data.UserId).Preload("UserDevices").Preload("AuthMethods").First(v).Error
	})
	return &user, err
}

func (m *defaultUserModel) Insert(ctx context.Context, data *User, tx ...*gorm.DB) error {
	err := m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(&data).Error
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultUserModel) FindOne(ctx context.Context, id int64) (*User, error) {
	userIdKey := fmt.Sprintf("%s%v", cacheUserIdPrefix, id)
	var resp User
	err := m.QueryCtx(ctx, &resp, userIdKey, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Where("`id` = ?", id).Preload("UserDevices").Preload("AuthMethods").First(&resp).Error
	})
	return &resp, err
}

func (m *defaultUserModel) Update(ctx context.Context, data *User, tx ...*gorm.DB) error {
	old, err := m.FindOne(ctx, data.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	err = m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Save(data).Error
	}, m.getCacheKeys(old)...)
	return err
}

func (m *defaultUserModel) Delete(ctx context.Context, id int64, tx ...*gorm.DB) error {
	data, err := m.FindOne(ctx, id)
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
		return conn.Transaction(func(db *gorm.DB) error {
			if err := db.Model(&User{}).Where("`id` = ?", id).Delete(&User{}).Error; err != nil {
				return err
			}
			if err := db.Model(&AuthMethods{}).Where("`user_id` = ?", id).Delete(&User{}).Error; err != nil {
				return err
			}
			if err := db.Model(&Subscribe{}).Where("`user_id` = ?", id).Delete(&User{}).Error; err != nil {
				return err
			}
			if err := db.Model(&BalanceLog{}).Where("`user_id` = ?", id).Delete(&User{}).Error; err != nil {
				return err
			}
			if err := db.Model(&GiftAmountLog{}).Where("`user_id` = ?", id).Delete(&User{}).Error; err != nil {
				return err
			}
			if err := db.Model(&LoginLog{}).Where("`user_id` = ?", id).Delete(&User{}).Error; err != nil {
				return err
			}
			if err := db.Model(&SubscribeLog{}).Where("`user_id` = ?", id).Delete(&User{}).Error; err != nil {
				return err
			}
			if err := db.Model(&Device{}).Where("`user_id` = ?", id).Delete(&User{}).Error; err != nil {
				return err
			}

			subs, err := m.QueryUserSubscribe(ctx, id)
			if err != nil {
				return err
			}
			for _, sub := range subs {
				if err := m.DeleteSubscribeById(ctx, sub.Id, db); err != nil {
					return err
				}
			}

			if err := db.Model(&CommissionLog{}).Where("`user_id` = ?", id).Delete(&User{}).Error; err != nil {
				return err
			}
			return nil
		})
	}, m.getCacheKeys(data)...)
	return err
}

func (m *defaultUserModel) Transaction(ctx context.Context, fn func(db *gorm.DB) error) error {
	return m.TransactCtx(ctx, fn)
}
