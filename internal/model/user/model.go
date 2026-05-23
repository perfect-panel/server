package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/pkg/orm"

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
	Note        string               `gorm:"type:varchar(500);default:'';comment:User note for subscription"`
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
	Unscoped        bool   // Whether to include soft-deleted records
}

type EmailRecipientFilter struct {
	Scope             int8
	RegisterStartTime int64
	RegisterEndTime   int64
}

type SubscribeFilter struct {
	Subscribers []int64
	IsActive    *bool
	StartTime   int64
	EndTime     int64
}

type customUserLogicModel interface {
	QueryPageList(ctx context.Context, page, size int, filter *UserFilterParams) ([]*User, int64, error)
	FindUsersByIds(ctx context.Context, ids []int64) ([]*User, error)
	CountAffiliates(ctx context.Context, refererId int64) (int64, error)
	QueryAffiliateList(ctx context.Context, refererId int64, page, size int) ([]*User, int64, error)
	FindOneByReferCode(ctx context.Context, referCode string) (*User, error)
	BatchDeleteUser(ctx context.Context, ids []int64, tx ...*gorm.DB) error
	InsertSubscribe(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error
	FindOneSubscribeByToken(ctx context.Context, token string) (*Subscribe, error)
	FindOneSubscribeByOrderId(ctx context.Context, orderId int64) (*Subscribe, error)
	FindOneSubscribe(ctx context.Context, id int64) (*Subscribe, error)
	FindSubscribesByIds(ctx context.Context, ids []int64) ([]*Subscribe, error)
	QueryMonthlyResetSubscribeIds(ctx context.Context, subscribeIds []int64, now time.Time) ([]int64, error)
	QueryFirstResetSubscribeIds(ctx context.Context, subscribeIds []int64) ([]int64, error)
	QueryYearlyResetSubscribeIds(ctx context.Context, subscribeIds []int64, now time.Time) ([]int64, error)
	ResetSubscribeTrafficByIds(ctx context.Context, ids []int64, tx ...*gorm.DB) error
	FindTrafficExceededSubscribes(ctx context.Context) ([]*Subscribe, error)
	FindExpiredSubscribes(ctx context.Context, now time.Time) ([]*Subscribe, error)
	MarkSubscribesFinished(ctx context.Context, ids []int64, status uint8, finishedAt time.Time, tx ...*gorm.DB) error
	UpdateSubscribe(ctx context.Context, data *Subscribe, tx ...*gorm.DB) error
	DeleteSubscribe(ctx context.Context, token string, tx ...*gorm.DB) error
	DeleteSubscribeById(ctx context.Context, id int64, tx ...*gorm.DB) error
	QueryUserSubscribe(ctx context.Context, userId int64, status ...int64) ([]*SubscribeDetails, error)
	FindOneSubscribeDetailsById(ctx context.Context, id int64) (*SubscribeDetails, error)
	FindOneUserSubscribe(ctx context.Context, id int64) (*SubscribeDetails, error)
	FindUsersSubscribeBySubscribeId(ctx context.Context, subscribeId int64) ([]*Subscribe, error)
	FindUserSubscribesByStatus(ctx context.Context, status ...int64) ([]*Subscribe, error)
	ActivatePendingSubscribesBySubscribeId(ctx context.Context, subscribeId int64) error
	CountUserSubscribesByUserAndSubscribe(ctx context.Context, userId, subscribeId int64) (int64, error)
	CountUserSubscribesBySubscribeIdAndStatus(ctx context.Context, subscribeId int64, status ...int64) (int64, error)
	UpdateUserSubscribeWithTraffic(ctx context.Context, id, download, upload int64, tx ...*gorm.DB) error
	QueryResisterUserTotalByDate(ctx context.Context, date time.Time) (int64, error)
	QueryResisterUserTotalByMonthly(ctx context.Context, date time.Time) (int64, error)
	QueryResisterUserTotal(ctx context.Context) (int64, error)
	CountEnabledUsers(ctx context.Context) (int64, error)
	QueryAdminUsers(ctx context.Context) ([]*User, error)
	UpdateUserCache(ctx context.Context, data *User) error
	UpdateUserSubscribeCache(ctx context.Context, data *Subscribe) error
	QueryActiveSubscriptions(ctx context.Context, subscribeId ...int64) (map[int64]int64, error)
	FindUserAuthMethods(ctx context.Context, userId int64) ([]*AuthMethods, error)
	InsertUserAuthMethods(ctx context.Context, data *AuthMethods, tx ...*gorm.DB) error
	UpdateUserAuthMethods(ctx context.Context, data *AuthMethods, tx ...*gorm.DB) error
	DeleteUserAuthMethods(ctx context.Context, userId int64, platform string, tx ...*gorm.DB) error
	UpdateUserAuthMethodOwner(ctx context.Context, authType, identifier string, userId int64, tx ...*gorm.DB) error
	DeleteUserAuthMethodByIdentifier(ctx context.Context, authType, identifier string, tx ...*gorm.DB) error
	UpsertUserAuthMethod(ctx context.Context, data *AuthMethods) error
	FindUserAuthMethodByOpenID(ctx context.Context, method, openID string) (*AuthMethods, error)
	FindUserAuthMethodByUserId(ctx context.Context, method string, userId int64) (*AuthMethods, error)
	FindUserAuthMethodByPlatform(ctx context.Context, userId int64, platform string) (*AuthMethods, error)
	QueryEmailRecipients(ctx context.Context, filter *EmailRecipientFilter) ([]string, error)
	CountEmailRecipients(ctx context.Context, filter *EmailRecipientFilter) (int64, error)
	FindOneByEmail(ctx context.Context, email string) (*User, error)
	FindOneDevice(ctx context.Context, id int64) (*Device, error)
	QueryDeviceList(ctx context.Context, userid int64) ([]*Device, int64, error)
	QueryDevicePageList(ctx context.Context, userid, subscribeId int64, page, size int) ([]*Device, int64, error)
	UpdateDevice(ctx context.Context, data *Device, tx ...*gorm.DB) error
	FindOneDeviceByIdentifier(ctx context.Context, id string) (*Device, error)
	DeleteDevice(ctx context.Context, id int64, tx ...*gorm.DB) error
	InsertDevice(ctx context.Context, data *Device, tx ...*gorm.DB) error
	FindDeviceOnlineRecord(ctx context.Context, userId int64, startTime, endTime string) (*DeviceOnlineRecord, error)
	InsertDeviceOnlineRecord(ctx context.Context, data *DeviceOnlineRecord, tx ...*gorm.DB) error
	InsertWithdrawal(ctx context.Context, data *Withdrawal, tx ...*gorm.DB) error

	QuerySubscribeIdsByFilter(ctx context.Context, filter *SubscribeFilter) ([]int64, error)
	CountSubscribesByFilter(ctx context.Context, filter *SubscribeFilter) (int64, error)

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
		conn = applyUserPageFilters(conn.Model(&User{}), filter)
		if err := conn.Count(&total).Error; err != nil {
			return err
		}
		return conn.Limit(size).Offset((page - 1) * size).Preload("UserDevices").Preload("AuthMethods").Find(&list).Error
	})
	return list, total, err
}

func applyUserPageFilters(conn *gorm.DB, filter *UserFilterParams) *gorm.DB {
	userIdColumn := UserColumn(conn, "id")
	if filter == nil {
		return conn
	}
	if filter.UserId != nil {
		conn = conn.Where(userIdColumn+" = ?", *filter.UserId)
	}
	if filter.Search != "" {
		search := orm.LikePrefixPattern(filter.Search)
		if search != "" {
			conn = conn.Where(userSearchCondition(conn), search, search)
		}
	}
	if filter.UserSubscribeId != nil {
		conn = conn.Where(
			fmt.Sprintf(
				"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s = ? AND %s IN ?)",
				UserSubscribeTableName(conn),
				UserSubscribeColumn(conn, "user_id"),
				userIdColumn,
				UserSubscribeColumn(conn, "id"),
				UserSubscribeColumn(conn, "status"),
			),
			*filter.UserSubscribeId,
			[]int64{0, 1},
		)
	}
	if filter.SubscribeId != nil {
		conn = conn.Where(
			fmt.Sprintf(
				"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s = ? AND %s IN ?)",
				UserSubscribeTableName(conn),
				UserSubscribeColumn(conn, "user_id"),
				userIdColumn,
				UserSubscribeColumn(conn, "subscribe_id"),
				UserSubscribeColumn(conn, "status"),
			),
			*filter.SubscribeId,
			[]int64{0, 1},
		)
	}
	if filter.Order != "" {
		switch strings.ToUpper(filter.Order) {
		case "ASC", "DESC":
			conn = conn.Order(fmt.Sprintf("%s %s", userIdColumn, strings.ToUpper(filter.Order)))
		}
	}
	if filter.Unscoped {
		conn = conn.Unscoped()
	}
	return conn
}

func userSearchCondition(conn *gorm.DB) string {
	return fmt.Sprintf(
		"(%s LIKE ?%s OR EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s LIKE ?%s))",
		UserColumn(conn, "refer_code"),
		orm.LikeEscapeClause(),
		AuthMethodsTableName(conn),
		AuthMethodsColumn(conn, "user_id"),
		UserColumn(conn, "id"),
		AuthMethodsColumn(conn, "auth_identifier"),
		orm.LikeEscapeClause(),
	)
}

func emailRecipientQuery(conn *gorm.DB, filter *EmailRecipientFilter) *gorm.DB {
	if filter == nil {
		filter = &EmailRecipientFilter{Scope: 1}
	}
	userID := UserColumn(conn, "id")
	userCreatedAt := UserColumn(conn, "created_at")
	authUserID := AuthMethodsColumn(conn, "user_id")
	authType := AuthMethodsColumn(conn, "auth_type")
	query := conn.Model(&AuthMethods{}).
		Select("auth_identifier").
		Joins(fmt.Sprintf("JOIN %s ON %s = %s", UserTableName(conn), userID, authUserID)).
		Where(authType+" = ?", "email")

	if filter.RegisterStartTime != 0 {
		query = query.Where(userCreatedAt+" >= ?", time.UnixMilli(filter.RegisterStartTime))
	}
	if filter.RegisterEndTime != 0 {
		query = query.Where(userCreatedAt+" <= ?", time.UnixMilli(filter.RegisterEndTime))
	}

	switch filter.Scope {
	case 2:
		query = query.Joins(fmt.Sprintf("JOIN user_subscribe ON %s = user_subscribe.user_id", userID)).
			Where("user_subscribe.status IN ?", []int64{1, 2})
	case 3:
		query = query.Joins(fmt.Sprintf("JOIN user_subscribe ON %s = user_subscribe.user_id", userID)).
			Where("user_subscribe.status = ?", 3)
	case 4:
		query = query.Joins(fmt.Sprintf("LEFT JOIN user_subscribe ON %s = user_subscribe.user_id", userID)).
			Where("user_subscribe.user_id IS NULL")
	}
	return query
}

func (m *customUserModel) QueryEmailRecipients(ctx context.Context, filter *EmailRecipientFilter) ([]string, error) {
	if filter != nil && filter.Scope == 5 {
		return nil, nil
	}
	var emails []string
	err := m.QueryNoCacheCtx(ctx, &emails, func(conn *gorm.DB, v interface{}) error {
		return emailRecipientQuery(conn, filter).Pluck("auth_identifier", v).Error
	})
	return emails, err
}

func (m *customUserModel) CountEmailRecipients(ctx context.Context, filter *EmailRecipientFilter) (int64, error) {
	if filter != nil && filter.Scope == 5 {
		return 0, nil
	}
	var total int64
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return emailRecipientQuery(conn, filter).Count(&total).Error
	})
	return total, err
}

func subscribeFilterQuery(conn *gorm.DB, filter *SubscribeFilter) *gorm.DB {
	query := conn.Model(&Subscribe{})
	if filter == nil {
		return query
	}
	if len(filter.Subscribers) > 0 {
		query = query.Where("subscribe_id IN ?", filter.Subscribers)
	}
	if filter.IsActive != nil && *filter.IsActive {
		query = query.Where("status IN ?", []int64{0, 1, 2})
	}
	if filter.StartTime != 0 {
		query = query.Where("start_time <= ?", time.UnixMilli(filter.StartTime))
	}
	if filter.EndTime != 0 {
		query = query.Where("expire_time >= ?", time.UnixMilli(filter.EndTime))
	}
	return query
}

func (m *customUserModel) QuerySubscribeIdsByFilter(ctx context.Context, filter *SubscribeFilter) ([]int64, error) {
	var ids []int64
	err := m.QueryNoCacheCtx(ctx, &ids, func(conn *gorm.DB, v interface{}) error {
		return subscribeFilterQuery(conn, filter).Pluck("id", v).Error
	})
	return ids, err
}

func (m *customUserModel) CountSubscribesByFilter(ctx context.Context, filter *SubscribeFilter) (int64, error) {
	var total int64
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return subscribeFilterQuery(conn, filter).Count(&total).Error
	})
	return total, err
}

// BatchDeleteUser deletes multiple records by primary key.
func (m *customUserModel) BatchDeleteUser(ctx context.Context, ids []int64, tx ...*gorm.DB) error {
	if len(ids) == 0 {
		return nil
	}
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
		if len(tx) > 0 {
			conn = tx[0]
		}
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
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.AddDate(0, 0, 1)
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&total).Error
	})
	return total, err
}

func (m *customUserModel) QueryResisterUserTotalByMonthly(ctx context.Context, date time.Time) (int64, error) {
	var total int64
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	end := start.AddDate(0, 1, 0)
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&total).Error
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

func (m *customUserModel) CountEnabledUsers(ctx context.Context) (int64, error) {
	var total int64
	err := m.QueryNoCacheCtx(ctx, &total, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&User{}).Where("enable = ?", true).Count(&total).Error
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
		orderDateExpr := userDateBucketExpr(conn, "created_at", "day")
		userCreatedAt := UserColumn(conn, "created_at")
		userDateExpr := userDateBucketExpr(conn, userCreatedAt, "day")

		// 子查询：统计每天的新用户订单数量
		newOrderSub := conn.Model(&order.Order{}).
			Select(fmt.Sprintf("%s AS date, COUNT(DISTINCT user_id) AS new_order_users", orderDateExpr)).
			Where("is_new = ? AND created_at BETWEEN ? AND ? AND status IN ?", true, firstDay, date, []int64{2, 5}).
			Group(orderDateExpr)

		// 子查询：统计每天的续费订单数量
		renewalOrderSub := conn.Model(&order.Order{}).
			Select(fmt.Sprintf("%s AS date, COUNT(DISTINCT user_id) AS renewal_order_users", orderDateExpr)).
			Where("is_new = ? AND created_at BETWEEN ? AND ? AND status IN ?", false, firstDay, date, []int64{2, 5}).
			Group(orderDateExpr)

		return conn.Model(&User{}).
			Select(fmt.Sprintf(`
                %s AS date,
                COUNT(*) AS register,
                COALESCE(MAX(n.new_order_users), 0) AS new_order_users,
                COALESCE(MAX(r.renewal_order_users), 0) AS renewal_order_users
            `, userDateExpr)).
			Joins("LEFT JOIN (?) AS n ON "+userDateExpr+" = n.date", newOrderSub).
			Joins("LEFT JOIN (?) AS r ON "+userDateExpr+" = r.date", renewalOrderSub).
			Where(userCreatedAt+" BETWEEN ? AND ?", firstDay, date).
			Group(userDateExpr).
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
		orderDateExpr := userDateBucketExpr(conn, "created_at", "month")
		userCreatedAt := UserColumn(conn, "created_at")
		userDateExpr := userDateBucketExpr(conn, userCreatedAt, "month")

		// 子查询：每月新订单用户数量
		newOrderSub := conn.Model(&order.Order{}).
			Select(fmt.Sprintf("%s AS date, COUNT(DISTINCT user_id) AS new_order_users", orderDateExpr)).
			Where("is_new = ? AND created_at >= ? AND status IN ?", true, sixMonthsAgo, []int64{2, 5}).
			Group(orderDateExpr)

		// 子查询：每月续费订单数量
		renewalOrderSub := conn.Model(&order.Order{}).
			Select(fmt.Sprintf("%s AS date, COUNT(DISTINCT user_id) AS renewal_order_users", orderDateExpr)).
			Where("is_new = ? AND created_at >= ? AND status IN ?", false, sixMonthsAgo, []int64{2, 5}).
			Group(orderDateExpr)

		return conn.Model(&User{}).
			Select(fmt.Sprintf(`
				%s AS date,
				COUNT(*) AS register,
				COALESCE(MAX(n.new_order_users), 0) AS new_order_users,
				COALESCE(MAX(r.renewal_order_users), 0) AS renewal_order_users
			`, userDateExpr)).
			Joins("LEFT JOIN (?) AS n ON "+userDateExpr+" = n.date", newOrderSub).
			Joins("LEFT JOIN (?) AS r ON "+userDateExpr+" = r.date", renewalOrderSub).
			Where(userCreatedAt+" >= ?", sixMonthsAgo).
			Group(userDateExpr).
			Order("date ASC").
			Scan(v).Error
	})

	return results, err
}

func userDateBucketExpr(db *gorm.DB, column, bucket string) string {
	if db.Dialector.Name() == "postgres" {
		if bucket == "month" {
			return fmt.Sprintf("TO_CHAR(%s, 'YYYY-MM')", column)
		}
		return fmt.Sprintf("TO_CHAR(%s, 'YYYY-MM-DD')", column)
	}
	if bucket == "month" {
		return fmt.Sprintf("DATE_FORMAT(%s, '%%Y-%%m')", column)
	}
	return fmt.Sprintf("DATE_FORMAT(%s, '%%Y-%%m-%%d')", column)
}
