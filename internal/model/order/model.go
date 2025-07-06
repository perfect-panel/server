package order

import (
	"context"
	"time"

	"github.com/perfect-panel/server/internal/model/payment"

	"github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Details struct {
	Id             int64                `gorm:"primaryKey"`
	ParentId       int64                `gorm:"type:bigint;default:null;comment:Parent Order Id"`
	SubOrders      []*Order             `gorm:"foreignKey:ParentId;references:Id"`
	UserId         int64                `gorm:"type:bigint;not null;default:0;comment:User Id"`
	OrderNo        string               `gorm:"type:varchar(255);not null;default:'';unique;comment:Order No"`
	Type           uint8                `gorm:"type:tinyint(1);not null;default:1;comment:Order Type: 1: Subscribe, 2: Renewal, 3: ResetTraffic, 4: Recharge"`
	Quantity       int64                `gorm:"type:bigint;not null;default:1;comment:Quantity"`
	Price          int64                `gorm:"type:int;not null;default:0;comment:Original price"`
	Amount         int64                `gorm:"type:int;not null;default:0;comment:Order Amount"`
	Discount       int64                `gorm:"type:int;not null;default:0;comment:Order Discount"`
	Coupon         string               `gorm:"type:varchar(255);default:null;comment:Coupon"`
	CouponDiscount int64                `gorm:"type:int;not null;default:0;comment:Coupon Discount"`
	PaymentId      int64                `gorm:"type:bigint;not null;default:0;comment:Payment Id"`
	Payment        *payment.Payment     `gorm:"foreignKey:PaymentId;references:Id"`
	Method         string               `gorm:"type:varchar(255);not null;default:'';comment:Payment Method"`
	FeeAmount      int64                `gorm:"type:int;not null;default:0;comment:Fee Amount"`
	TradeNo        string               `gorm:"type:varchar(255);default:null;comment:Trade No"`
	GiftAmount     int64                `gorm:"type:int;not null;default:0;comment:User Gift Amount"`
	Commission     int64                `gorm:"type:int;not null;default:0;comment:Order Commission"`
	Status         uint8                `gorm:"type:tinyint(1);not null;default:1;comment:Order Status: 1: Pending, 2: Paid, 3: Failed"`
	SubscribeId    int64                `gorm:"type:bigint;not null;default:0;comment:Subscribe Id"`
	SubscribeToken string               `gorm:"type:varchar(255);default:null;comment:Renewal Subscribe Token"`
	Subscribe      *subscribe.Subscribe `gorm:"foreignKey:SubscribeId;references:Id"`
	IsNew          bool                 `gorm:"type:tinyint(1);not null;default:0;comment:Is New Order"`
	CreatedAt      time.Time            `gorm:"<-:create;comment:Create Time"`
	UpdatedAt      time.Time            `gorm:"comment:Update Time"`
}

type customOrderLogicModel interface {
	UpdateOrderStatus(ctx context.Context, orderNo string, status uint8, tx ...*gorm.DB) error
	QueryOrderListByPage(ctx context.Context, page, size int, status uint8, user, subscribe int64, search string) (int64, []*Details, error)
	FindOneDetails(ctx context.Context, id int64) (*Details, error)
	FindOneDetailsByOrderNo(ctx context.Context, orderNo string) (*Details, error)
	QueryMonthlyOrders(ctx context.Context, date time.Time) (OrdersTotal, error)
	QueryDateOrders(ctx context.Context, date time.Time) (OrdersTotal, error)
	QueryTotalOrders(ctx context.Context) (OrdersTotal, error)
	QueryMonthlyUserCounts(ctx context.Context, date time.Time) (int64, int64, error)
	QueryDateUserCounts(ctx context.Context, date time.Time) (int64, int64, error)
	QueryTotalUserCounts(ctx context.Context) (int64, int64, error)
	IsUserEligibleForNewOrder(ctx context.Context, userID int64) (bool, error)
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customOrderModel{
		defaultOrderModel: newOrderModel(conn, c),
	}
}

// QueryOrderListByPage Query order list by page
func (m *customOrderModel) QueryOrderListByPage(ctx context.Context, page, size int, status uint8, user, subscribe int64, search string) (int64, []*Details, error) {
	var list []*Details
	var total int64
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		conn = conn.Model(&Order{})
		if status > 0 {
			conn = conn.Where("status = ?", status)
		}
		if user > 0 {
			conn = conn.Where("user_id = ?", user)
		}
		if subscribe > 0 {
			conn = conn.Where("subscribe_id = ?", subscribe)
		}
		if search != "" {
			conn = conn.Where("order_no like ? or trade_no like ? or coupon like ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
		}
		return conn.Order("id desc").Preload("Subscribe").Preload("Payment").Count(&total).Offset((page - 1) * size).Limit(size).Find(v).Error
	})
	return total, list, err
}

// UpdateOrderStatus Update order status
func (m *customOrderModel) UpdateOrderStatus(ctx context.Context, orderNo string, status uint8, tx ...*gorm.DB) error {
	orderInfo, err := m.FindOneByOrderNo(ctx, orderNo)
	if err != nil {
		return err
	}
	return m.ExecCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Model(&Order{}).Where("order_no = ?", orderNo).Update("status", status).Error
	}, m.getCacheKeys(orderInfo)...)
}

// FindOneDetailsByOrderNo Find order details by order number
func (m *customOrderModel) FindOneDetailsByOrderNo(ctx context.Context, orderNo string) (*Details, error) {
	var orderInfo Details
	err := m.QueryNoCacheCtx(ctx, &orderInfo, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Order{}).Where("order_no = ?", orderNo).Preload("Subscribe").Preload("Payment").First(v).Error
	})
	return &orderInfo, err
}

func (m *customOrderModel) FindOneDetails(ctx context.Context, id int64) (*Details, error) {
	var orderInfo Details
	err := m.QueryNoCacheCtx(ctx, &orderInfo, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Order{}).
			Where("id = ?", id).
			Preload("Subscribe").
			Preload("SubOrders").
			First(v).Error
	})
	return &orderInfo, err
}

func (m *customOrderModel) QueryMonthlyOrders(ctx context.Context, date time.Time) (OrdersTotal, error) {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Nanosecond)
	var result OrdersTotal
	err := m.QueryNoCacheCtx(ctx, &result, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Order{}).
			Where("status IN ? AND created_at BETWEEN ? AND ? AND method != ?", []int64{2, 5}, firstDay, lastDay, "balance").
			Select(
				"SUM(amount) as amount_total, " +
					"SUM(CASE WHEN is_new = 1 THEN amount ELSE 0 END) as new_order_amount, " +
					"SUM(CASE WHEN is_new = 0 THEN amount ELSE 0 END) as renewal_order_amount",
			).
			Scan(v).Error
	})
	return result, err
}

// QueryDateOrders Query orders by date
func (m *customOrderModel) QueryDateOrders(ctx context.Context, date time.Time) (OrdersTotal, error) {
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)
	var result OrdersTotal
	err := m.QueryNoCacheCtx(ctx, &result, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Order{}).
			Where("status IN ? AND created_at BETWEEN ? AND ? AND method != ?", []int64{2, 5}, start, end, "balance").
			Select(
				"SUM(amount) as amount_total, " +
					"SUM(CASE WHEN is_new = 1 THEN amount ELSE 0 END) as new_order_amount, " +
					"SUM(CASE WHEN is_new = 0 THEN amount ELSE 0 END) as renewal_order_amount",
			).
			Scan(v).Error
	})
	return result, err
}

func (m *customOrderModel) QueryTotalOrders(ctx context.Context) (OrdersTotal, error) {
	var result OrdersTotal
	err := m.QueryNoCacheCtx(ctx, &result, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&Order{}).
			Where("status IN ? AND method != ?", []int64{2, 5}, "balance").
			Select(
				"SUM(amount) as amount_total, " +
					"SUM(CASE WHEN is_new = 1 THEN amount ELSE 0 END) as new_order_amount, " +
					"SUM(CASE WHEN is_new = 0 THEN amount ELSE 0 END) as renewal_order_amount",
			).
			Scan(v).Error
	})
	return result, err
}

func (m *customOrderModel) QueryMonthlyUserCounts(ctx context.Context, date time.Time) (int64, int64, error) {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	var newUsers int64
	var renewalUsers int64
	err := m.QueryNoCacheCtx(ctx, nil, func(conn *gorm.DB, _ interface{}) error {
		return conn.Model(&Order{}).
			Where("status IN ? AND created_at BETWEEN ? AND ? AND method != ?", []int64{2, 5}, firstDay, lastDay, "balance").
			Select(
				"COUNT(DISTINCT CASE WHEN is_new = 1 THEN user_id END) as new_users, "+
					"COUNT(DISTINCT CASE WHEN is_new = 0 THEN user_id END) as renewal_users").
			Row().Scan(&newUsers, &renewalUsers)
	})
	return newUsers, renewalUsers, err
}

func (m *customOrderModel) QueryDateUserCounts(ctx context.Context, date time.Time) (int64, int64, error) {
	start := date.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour).Add(-time.Nanosecond)

	var newUsers int64
	var renewalUsers int64
	err := m.QueryNoCacheCtx(ctx, nil, func(conn *gorm.DB, _ interface{}) error {
		return conn.Model(&Order{}).
			Where("status IN ? AND created_at BETWEEN ? AND ? AND method != ?", []int64{2, 5}, start, end, "balance").
			Select(
				"COUNT(DISTINCT CASE WHEN is_new = 1 THEN user_id END) as new_users, "+
					"COUNT(DISTINCT CASE WHEN is_new = 0 THEN user_id END) as renewal_users").
			Row().Scan(&newUsers, &renewalUsers)
	})
	return newUsers, renewalUsers, err
}

func (m *customOrderModel) QueryTotalUserCounts(ctx context.Context) (int64, int64, error) {
	var newUsers int64
	var renewalUsers int64
	err := m.QueryNoCacheCtx(ctx, nil, func(conn *gorm.DB, _ interface{}) error {
		return conn.Model(&Order{}).
			Where("status IN ? AND method != ?", []int64{2, 5}, "balance").
			Select(
				"COUNT(DISTINCT CASE WHEN is_new = 1 THEN user_id END) as new_users, "+
					"COUNT(DISTINCT CASE WHEN is_new = 0 THEN user_id END) as renewal_users").
			Row().Scan(&newUsers, &renewalUsers)
	})
	return newUsers, renewalUsers, err
}

func (m *customOrderModel) IsUserEligibleForNewOrder(ctx context.Context, userID int64) (bool, error) {
	var count int64
	err := m.QueryNoCacheCtx(ctx, nil, func(conn *gorm.DB, _ interface{}) error {
		return conn.Model(&Order{}).
			Where("user_id = ? AND status IN ?", userID, []int64{2, 5}).
			Count(&count).Error
	})
	return count == 0, err
}
