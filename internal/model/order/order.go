package order

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Order struct {
	Id             int64     `gorm:"primaryKey"`
	ParentId       int64     `gorm:"type:bigint;default:null;comment:Parent Order Id"`
	UserId         int64     `gorm:"type:bigint;not null;default:0;comment:User Id"`
	OrderNo        string    `gorm:"type:varchar(255);not null;default:'';unique;comment:Order No"`
	Type           uint8     `gorm:"type:tinyint(1);not null;default:1;comment:Order Type: 1: Subscribe, 2: Renewal, 3: ResetTraffic, 4: Recharge"`
	Quantity       int64     `gorm:"type:bigint;not null;default:1;comment:Quantity"`
	Price          int64     `gorm:"type:int;not null;default:0;comment:Original price"`
	Amount         int64     `gorm:"type:int;not null;default:0;comment:Order Amount"`
	GiftAmount     int64     `gorm:"type:int;not null;default:0;comment:User Gift Amount"`
	Discount       int64     `gorm:"type:int;not null;default:0;comment:Discount Amount"`
	Coupon         string    `gorm:"type:varchar(255);default:null;comment:Coupon"`
	CouponDiscount int64     `gorm:"type:int;not null;default:0;comment:Coupon Discount Amount"`
	Commission     int64     `gorm:"type:int;not null;default:0;comment:Order Commission"`
	PaymentId      int64     `gorm:"type:bigint;not null;default:0;comment:Payment Method Id"`
	Method         string    `gorm:"type:varchar(255);not null;default:'';comment:Payment Method"`
	FeeAmount      int64     `gorm:"type:int;not null;default:0;comment:Fee Amount"`
	TradeNo        string    `gorm:"type:varchar(255);default:null;comment:Trade No"`
	Status         uint8     `gorm:"type:tinyint(1);not null;default:1;comment:Order Status: 1: Pending, 2: Paid, 3:Close, 4: Failed, 5:Finished;"`
	SubscribeId    int64     `gorm:"type:bigint;not null;default:0;comment:Subscribe Id"`
	SubscribeToken string    `gorm:"type:varchar(255);default:null;comment:Renewal Subscribe Token"`
	IsNew          bool      `gorm:"type:tinyint(1);not null;default:0;comment:Is New Order"`
	CreatedAt      time.Time `gorm:"<-:create;comment:Create Time"`
	UpdatedAt      time.Time `gorm:"comment:Update Time"`
}

type OrdersTotal struct {
	AmountTotal        int64
	NewOrderAmount     int64
	RenewalOrderAmount int64
}

func (Order) TableName() string {
	return "order"
}

func orderTableName(db *gorm.DB) string {
	return quoteTable(db, Order{}.TableName())
}

func orderColumn(db *gorm.DB, column string) string {
	return quoteColumn(db, Order{}.TableName(), column)
}

func quoteTable(db *gorm.DB, table string) string {
	if db != nil && db.Statement != nil {
		return db.Statement.Quote(clause.Table{Name: table})
	}
	return table
}

func quoteColumn(db *gorm.DB, table, column string) string {
	if db != nil && db.Statement != nil {
		return db.Statement.Quote(clause.Column{Table: table, Name: column})
	}
	return table + "." + column
}
