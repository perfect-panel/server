package user

import (
	"context"
	"fmt"
	"time"

	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/subscribe"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	cacheUserSubscribeTokenPrefix = "cache:user:subscribe:token:"
	cacheUserSubscribeUserPrefix  = "cache:user:subscribe:user:"
	cacheUserSubscribeIdPrefix    = "cache:user:subscribe:id:"
	cacheUserDeviceNumberPrefix   = "cache:user:device:number:"
	cacheUserDeviceIdPrefix       = "cache:user:device:id:"
)

type SubscribeDetails struct {
	Id          int64                `gorm:"primarykey"`
	UserId      int64                `gorm:"index:idx_user_id;not null;comment:User ID"`
	User        *User                `gorm:"foreignKey:UserId;references:Id"`
	OrderId     int64                `gorm:"index:idx_order_id;not null;comment:Order ID"`
	SubscribeId int64                `gorm:"index:idx_subscribe_id;not null;comment:Subscription ID"`
	Subscribe   *subscribe.Subscribe `gorm:"foreignKey:SubscribeId;references:Id"`
	StartTime   time.Time            `gorm:"default:CURRENT_TIMESTAMP(3);not null;comment:Subscription Start Time"`
	ExpireTime  time.Time            `gorm:"default:NULL;comment:Subscription Expire Time"`
	FinishedAt  *time.Time           `gorm:"default:NULL;comment:Finished Time"`
	Traffic     int64                `gorm:"default:0;comment:Traffic"`
	Download    int64                `gorm:"default:0;comment:Download Traffic"`
	Upload      int64                `gorm:"default:0;comment:Upload Traffic"`
	Token       string               `gorm:"index:idx_token;unique;type:varchar(255);default:'';comment:Token"`
	UUID        string               `gorm:"type:varchar(255);unique;index:idx_uuid;default:'';comment:UUID"`
	Status      uint8                `gorm:"type:tinyint(1);default:0;comment:Subscription Status: 0: Pending 1: Active 2: Finished 3: Expired; 4: Cancelled"`
	CreatedAt   time.Time            `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt   time.Time            `gorm:"comment:Update Time"`
}

type SubscribeLogFilterParams struct {
	IP              string
	UserAgent       string
	UserId          int64
	Token           string
	UserSubscribeId int64
}

type LoginLogFilterParams struct {
	IP        string
	UserId    int64
	UserAgent string
	Success   *bool
}

type UserFilterParams struct {
	Search          string
	UserId          *int64
	SubscribeId     *int64
	UserSubscribeId *int64
	Order           string // Order by id, e.g., "desc"
}

type customUserLogicModel interface {
	QueryPageList(ctx context.Context, page, size int, filter *UserFilterParams) ([]*User, int64, error)
	FindOneByReferCode(ctx context.Context, referCode string) (*User, error)
	BatchDeleteUser(ctx context.Context, ids []int64, tx ...*gorm.DB) error
	InsertSubscribe(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error
	FindOneSubscribeByToken(ctx context.Context, token string) (*Subscribe, error)
	FindOneSubscribeByOrderId(ctx context.Context, orderId int64) (*Subscribe, error)
	FindOneSubscribe(ctx context.Context, id int64) (*Subscribe, error)
	UpdateSubscribe(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error
	DeleteSubscribe(ctx context.Context, token string, tx ...*gorm.DB) error
	DeleteSubscribeById(ctx context.Context, id int64, tx ...*gorm.DB) error
	QueryUserSubscribe(ctx context.Context, userId int64, status ...int64) ([]*SubscribeDetails, error)
	FindOneSubscribeDetailsById(ctx context.Context, id int64) (*SubscribeDetails, error)
	FindOneUserSubscribe(ctx context.Context, id int64) (*SubscribeDetails, error)
	FindUsersSubscribeBySubscribeId(ctx context.Context, subscribeId int64) ([]*Subscribe, error)
	UpdateUserSubscribeWithTraffic(ctx context.Context, id, download, upload int64, tx ...*gorm.DB) error
	QueryResisterUserTotalByDate(ctx context.Context, date time.Time) (int64, error)
	QueryResisterUserTotalByMonthly(ctx context.Context, date time.Time) (int64, error)
	QueryResisterUserTotal(ctx context.Context) (int64, error)
	QueryAdminUsers(ctx context.Context) ([]*User, error)
	UpdateUserCache(ctx context.Context, data *User) error
	UpdateUserSubscribeCache(ctx context.Context, data *Subscribe) error
	QueryActiveSubscriptions(ctx context.Context, subscribeId ...int64) (map[int64]int64, error)
	FindUserAuthMethods(ctx context.Context, userId int64) ([]*AuthMethods, error)
	InsertUserAuthMethods(ctx context.Context, data *AuthMethods, tx ...*gorm.DB) error
	UpdateUserAuthMethods(ctx context.Context, data *AuthMethods, tx ...*gorm.DB) error
	DeleteUserAuthMethods(ctx context.Context, userId int64, platform string, tx ...*gorm.DB) error
	FindUserAuthMethodByOpenID(ctx context.Context, method, openID string) (*AuthMethods, error)
	FindUserAuthMethodByUserId(ctx context.Context, method string, userId int64) (*AuthMethods, error)
	FindUserAuthMethodByPlatform(ctx context.Context, userId int64, platform string) (*AuthMethods, error)
	FindOneByEmail(ctx context.Context, email string) (*User, error)
	FindOneDevice(ctx context.Context, id int64) (*Device, error)
	QueryDeviceList(ctx context.Context, userid int64) ([]*Device, int64, error)
	QueryDevicePageList(ctx context.Context, userid, subscribeId int64, page, size int) ([]*Device, int64, error)
	UpdateDevice(ctx context.Context, data *Device, tx ...*gorm.DB) error
	FindOneDeviceByIdentifier(ctx context.Context, id string) (*Device, error)
	DeleteDevice(ctx context.Context, id int64, tx ...*gorm.DB) error
	InsertDevice(ctx context.Context, data *Device, tx ...*gorm.DB) error

	ClearSubscribeCache(ctx context.Context, data ...*Subscribe) error
	ClearUserCache(ctx context.Context, data ...*User) error

	QueryDailyUserStatisticsList(ctx context.Context, date time.Time) ([]UserStatisticsWithDate, error)
	QueryMonthlyUserStatisticsList(ctx context.Context, date time.Time) ([]UserStatisticsWithDate, error)
}

type UserStatisticsWithDate struct {
	Date              string
	Register          int64
	NewOrderUsers     int64
	RenewalOrderUsers int64
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customUserModel{
		defaultUserModel: newUserModel(conn, c),
	}
}

// QueryPageList returns a list of records that meet the conditions.
func (m *customUserModel) QueryPageList(ctx context.Context, page, size int, filter *UserFilterParams) ([]*User, int64, error) {
	var list []*User
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		if filter != nil {
			if filter.UserId != nil {
				conn = conn.Where("user.id =?", *filter.UserId)
			}
			if filter.Search != "" {
				conn = conn.Joins("LEFT JOIN user_auth_methods ON user.id = user_auth_methods.user_id").
					Where("user_auth_methods.auth_identifier LIKE ?", "%"+filter.Search+"%").Or("user.refer_code like ?", "%"+filter.Search+"%")
			}
			if filter.UserSubscribeId != nil {
				conn = conn.Joins("LEFT JOIN user_subscribe ON user.id = user_subscribe.user_id").
					Where("user_subscribe.id =? and `status` IN (0,1)", *filter.UserSubscribeId)
			}
			if filter.SubscribeId != nil {
				conn = conn.Joins("LEFT JOIN user_subscribe ON user.id = user_subscribe.user_id").
					Where("user_subscribe.subscribe_id =? and `status` IN (0,1)", *filter.SubscribeId)
			}
			if filter.Order != "" {
				conn = conn.Order(fmt.Sprintf("user.id %s", filter.Order))
			}
		}
		return conn.Model(&User{}).Group("user.id").Count(&total).Limit(size).Offset((page - 1) * size).Preload("UserDevices").Preload("AuthMethods").Find(&list).Error
	})
	return list, total, err
}

// BatchDeleteUser deletes multiple records by primary key.
func (m *customUserModel) BatchDeleteUser(ctx context.Context, ids []int64, tx ...*gorm.DB) error {
	var users []*User
	err := m.QueryNoCacheCtx(ctx, &users, func(conn *gorm.DB, v interface{}) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Where("id in ?", ids).Find(&users).Error
	})
	if err != nil {
		return err
	}
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		return conn.Where("id in ?", ids).Delete(&User{}).Error
	}, m.batchGetCacheKeys(users...)...)
}

func (m *customUserModel) UpdateUserSubscribeWithTraffic(ctx context.Context, id, download, upload int64, tx ...*gorm.DB) error {
	sub, err := m.FindOneSubscribe(ctx, id)
	if err != nil {
		return err
	}

	// 使用 defer 确保更新后清理缓存
	defer func() {
		if clearErr := m.ClearSubscribeCacheByModels(ctx, sub); clearErr != nil {
			// 记录清理缓存错误
		}
	}()

	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Model(&Subscribe{}).Where("id = ?", id).Updates(map[string]interface{}{
			"download": gorm.Expr("download + ?", download),
			"upload":   gorm.Expr("upload + ?", upload),
		}).Error
	})
}

func (m *customUserModel) QueryResisterUserTotalByDate(ctx context.Context, date time.Time) (int64, error) {
	var total int64
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour).Add(-time.Second)
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Where("created_at > ? and created_at < ?", start, end).Count(&total).Error
	})
	return total, err
}

func (m *customUserModel) QueryResisterUserTotalByMonthly(ctx context.Context, date time.Time) (int64, error) {
	var total int64
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Where("created_at > ? and created_at < ?", start, end).Count(&total).Error
	})
	return total, err
}

func (m *customUserModel) QueryResisterUserTotal(ctx context.Context) (int64, error) {
	var total int64
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Count(&total).Error
	})
	return total, err
}

func (m *customUserModel) QueryAdminUsers(ctx context.Context) ([]*User, error) {
	var data []*User
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Preload("AuthMethods").Where("is_admin = ?", true).Find(&data).Error
	})
	return data, err
}

func (m *customUserModel) UpdateUserCache(ctx context.Context, data *User) error {
	return m.ClearUserCache(ctx, data)
}

func (m *customUserModel) FindOneByReferCode(ctx context.Context, referCode string) (*User, error) {
	var data User
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Where("refer_code = ?", referCode).First(&data).Error
	})
	return &data, err
}

func (m *customUserModel) FindOneSubscribeDetailsById(ctx context.Context, id int64) (*SubscribeDetails, error) {
	var data SubscribeDetails
	err := m.QueryNoCacheCtx(ctx, &data, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Subscribe{}).Preload("Subscribe").Preload("User").Where("id = ?", id).First(&data).Error
	})
	return &data, err
}

// QueryDailyUserStatisticsList Query daily user statistics list for the current month (from 1st to current date)
func (m *customUserModel) QueryDailyUserStatisticsList(ctx context.Context, date time.Time) ([]UserStatisticsWithDate, error) {
	var results []UserStatisticsWithDate

	err := m.QueryNoCacheCtx(ctx, &results, func(conn *gorm.DB, v interface{}) error {
		firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())

		// 子查询：统计每天的新用户订单数量
		newOrderSub := conn.Model(&order.Order{}).
			Select("DATE_FORMAT(created_at, '%Y-%m-%d') AS date, COUNT(DISTINCT user_id) AS new_order_users").
			Where("is_new = 1 AND created_at BETWEEN ? AND ? AND status IN ?", firstDay, date, []int64{2, 5}).
			Group("DATE_FORMAT(created_at, '%Y-%m-%d')")

		// 子查询：统计每天的续费订单数量
		renewalOrderSub := conn.Model(&order.Order{}).
			Select("DATE_FORMAT(created_at, '%Y-%m-%d') AS date, COUNT(DISTINCT user_id) AS renewal_order_users").
			Where("is_new = 0 AND created_at BETWEEN ? AND ? AND status IN ?", firstDay, date, []int64{2, 5}).
			Group("DATE_FORMAT(created_at, '%Y-%m-%d')")

		return conn.Model(&User{}).
			Select(`
                DATE_FORMAT(user.created_at, '%Y-%m-%d') AS date,
                COUNT(*) AS register,
                IFNULL(MAX(n.new_order_users), 0) AS new_order_users,
                IFNULL(MAX(r.renewal_order_users), 0) AS renewal_order_users
            `).
			Joins("LEFT JOIN (?) AS n ON DATE_FORMAT(user.created_at, '%Y-%m-%d') = n.date", newOrderSub).
			Joins("LEFT JOIN (?) AS r ON DATE_FORMAT(user.created_at, '%Y-%m-%d') = r.date", renewalOrderSub).
			Where("user.created_at BETWEEN ? AND ?", firstDay, date).
			Group("DATE_FORMAT(user.created_at, '%Y-%m-%d')").
			Order("date ASC").
			Scan(v).Error
	})

	return results, err
}

// QueryMonthlyUserStatisticsList Query monthly user statistics list for the past 6 months
func (m *customUserModel) QueryMonthlyUserStatisticsList(ctx context.Context, date time.Time) ([]UserStatisticsWithDate, error) {
	var results []UserStatisticsWithDate

	err := m.QueryNoCacheCtx(ctx, &results, func(conn *gorm.DB, v interface{}) error {
		// 获取 6 个月前的日期
		sixMonthsAgo := date.AddDate(0, -5, 0)

		// 子查询：每月新订单用户数量
		newOrderSub := conn.Model(&order.Order{}).
			Select("DATE_FORMAT(created_at, '%Y-%m') AS date, COUNT(DISTINCT user_id) AS new_order_users").
			Where("is_new = 1 AND created_at >= ? AND status IN ?", sixMonthsAgo, []int64{2, 5}).
			Group("DATE_FORMAT(created_at, '%Y-%m')")

		// 子查询：每月续费订单用户数量
		renewalOrderSub := conn.Model(&order.Order{}).
			Select("DATE_FORMAT(created_at, '%Y-%m') AS date, COUNT(DISTINCT user_id) AS renewal_order_users").
			Where("is_new = 0 AND created_at >= ? AND status IN ?", sixMonthsAgo, []int64{2, 5}).
			Group("DATE_FORMAT(created_at, '%Y-%m')")

		return conn.Model(&User{}).
			Select(`
				DATE_FORMAT(user.created_at, '%Y-%m') AS date,
				COUNT(*) AS register,
				IFNULL(MAX(n.new_order_users), 0) AS new_order_users,
				IFNULL(MAX(r.renewal_order_users), 0) AS renewal_order_users
			`).
			Joins("LEFT JOIN (?) AS n ON DATE_FORMAT(user.created_at, '%Y-%m') = n.date", newOrderSub).
			Joins("LEFT JOIN (?) AS r ON DATE_FORMAT(user.created_at, '%Y-%m') = r.date", renewalOrderSub).
			Where("user.created_at >= ?", sixMonthsAgo).
			Group("DATE_FORMAT(user.created_at, '%Y-%m')").
			Order("date ASC").
			Scan(v).Error
	})

	return results, err
}
