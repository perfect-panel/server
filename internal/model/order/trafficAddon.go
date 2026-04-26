package order

// V4.3 traffic_addon_order — 流量加购订单(决策 7)。
// 直接挂 order 包是因为它本质是一种订单流水,共用 Order 包的 GORM 入口更轻。

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type TrafficAddonOrder struct {
	Id              int64     `gorm:"primaryKey"`
	UserId          int64     `gorm:"column:user_id;index:idx_user_id;not null"`
	UserSubscribeId int64     `gorm:"column:user_subscribe_id;index:idx_user_subscribe_id;not null"`
	OrderId         int64     `gorm:"column:order_id;index:idx_order_id;not null;default:0;comment:Linked order id (0 if balance-paid only)"`
	AddonBytes      int64     `gorm:"column:addon_bytes;not null;comment:Bytes added to pool"`
	Amount          int64     `gorm:"column:amount;not null;comment:Paid amount in cents"`
	UnitPrice       int64     `gorm:"column:unit_price;not null;default:0;comment:Snapshot at purchase"`
	UnitSize        int64     `gorm:"column:unit_size;not null;default:0;comment:Snapshot at purchase"`
	CreatedAt       time.Time `gorm:"<-:create"`
	UpdatedAt       time.Time
}

func (*TrafficAddonOrder) TableName() string {
	return "traffic_addon_order"
}

// InsertTrafficAddonOrder 用于在加购流量包时记录订单流水。
func (m *defaultOrderModel) InsertTrafficAddonOrder(ctx context.Context, data *TrafficAddonOrder, tx ...*gorm.DB) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(data).Error
	})
}

// QueryTrafficAddonOrders 列出某用户订阅的加购流水。
func (m *defaultOrderModel) QueryTrafficAddonOrders(ctx context.Context, userSubscribeId int64) ([]*TrafficAddonOrder, error) {
	var list []*TrafficAddonOrder
	err := m.QueryNoCacheCtx(ctx, &list, func(conn *gorm.DB, v interface{}) error {
		return conn.Model(&TrafficAddonOrder{}).
			Where("user_subscribe_id = ?", userSubscribeId).
			Order("id DESC").
			Find(&list).Error
	})
	return list, err
}
