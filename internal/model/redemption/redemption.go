package redemption

import (
	"time"

	"gorm.io/gorm"
)

type RedemptionCode struct {
	Id            int64          `gorm:"primaryKey"`
	Code          string         `gorm:"type:varchar(255);not null;unique;comment:Redemption Code"`
	TotalCount    int64          `gorm:"type:int;not null;default:0;comment:Total Redemption Count"`
	UsedCount     int64          `gorm:"type:int;not null;default:0;comment:Used Redemption Count"`
	SubscribePlan int64          `gorm:"type:bigint;not null;default:0;comment:Subscribe Plan"`
	UnitTime      string         `gorm:"type:varchar(50);not null;default:'month';comment:Unit Time: day, month, quarter, half_year, year"`
	Quantity      int64          `gorm:"type:int;not null;default:1;comment:Quantity"`
	Status        int64          `gorm:"type:tinyint;not null;default:1;comment:Status: 1=enabled, 0=disabled"`
	CreatedAt     time.Time      `gorm:"<-:create;comment:Create Time"`
	UpdatedAt     time.Time      `gorm:"comment:Update Time"`
	DeletedAt     gorm.DeletedAt `gorm:"index;comment:Delete Time"`
}

type RedemptionRecord struct {
	Id               int64     `gorm:"primaryKey"`
	RedemptionCodeId int64     `gorm:"type:bigint;not null;default:0;comment:Redemption Code Id;index"`
	UserId           int64     `gorm:"type:bigint;not null;default:0;comment:User Id;index"`
	SubscribeId      int64     `gorm:"type:bigint;not null;default:0;comment:Subscribe Id"`
	UnitTime         string    `gorm:"type:varchar(50);not null;default:'month';comment:Unit Time"`
	Quantity         int64     `gorm:"type:int;not null;default:1;comment:Quantity"`
	RedeemedAt       time.Time `gorm:"<-:create;comment:Redeemed Time"`
	CreatedAt        time.Time `gorm:"<-:create;comment:Create Time"`
}

func (RedemptionCode) TableName() string {
	return "redemption_code"
}

func (RedemptionRecord) TableName() string {
	return "redemption_record"
}
