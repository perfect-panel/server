package subscribe

import (
	"time"

	"gorm.io/gorm"
)

type Subscribe struct {
	Id             int64     `gorm:"primaryKey"`
	Name           string    `gorm:"type:varchar(255);not null;default:'';comment:Subscribe Name"`
	Description    string    `gorm:"type:text;comment:Subscribe Description"`
	UnitPrice      int64     `gorm:"type:int;not null;default:0;comment:Unit Price"`
	UnitTime       string    `gorm:"type:varchar(255);not null;default:'';comment:Unit Time"`
	Discount       string    `gorm:"type:text;comment:Discount"`
	Replacement    int64     `gorm:"type:int;not null;default:0;comment:Replacement"`
	Inventory      int64     `gorm:"type:int;not null;default:0;comment:Inventory"`
	Traffic        int64     `gorm:"type:int;not null;default:0;comment:Traffic"`
	SpeedLimit     int64     `gorm:"type:int;not null;default:0;comment:Speed Limit"`
	DeviceLimit    int64     `gorm:"type:int;not null;default:0;comment:Device Limit"`
	Quota          int64     `gorm:"type:int;not null;default:0;comment:Quota"`
	GroupId        int64     `gorm:"type:bigint;comment:Group Id"`
	ServerGroup    string    `gorm:"type:varchar(255);comment:Server Group"`
	Server         string    `gorm:"type:varchar(255);comment:Server"`
	Show           *bool     `gorm:"type:tinyint(1);not null;default:0;comment:Show portal page"`
	Sell           *bool     `gorm:"type:tinyint(1);not null;default:0;comment:Sell"`
	Sort           int64     `gorm:"type:int;not null;default:0;comment:Sort"`
	DeductionRatio int64     `gorm:"type:int;default:0;comment:Deduction Ratio"`
	AllowDeduction *bool     `gorm:"type:tinyint(1);default:1;comment:Allow deduction"`
	ResetCycle     int64     `gorm:"type:int;default:0;comment:Reset Cycle: 0: No Reset, 1: 1st, 2: Monthly, 3: Yearly"`
	RenewalReset   *bool     `gorm:"type:tinyint(1);default:0;comment:Renew Reset"`
	CreatedAt      time.Time `gorm:"<-:create;comment:Create Time"`
	UpdatedAt      time.Time `gorm:"comment:Update Time"`
}

func (*Subscribe) TableName() string {
	return "subscribe"
}

func (s *Subscribe) BeforeCreate(tx *gorm.DB) error {
	if s.Sort == 0 {
		var maxSort int64
		if err := tx.Model(&Subscribe{}).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort).Error; err != nil {
			return err
		}
		s.Sort = maxSort + 1
	}
	return nil
}

type Discount struct {
	Months   int64 `json:"months"`
	Discount int64 `json:"discount"`
}

type Group struct {
	Id          int64     `gorm:"primaryKey"`
	Name        string    `gorm:"type:varchar(255);not null;default:'';comment:Group Name"`
	Description string    `gorm:"type:text;comment:Group Description"`
	CreatedAt   time.Time `gorm:"<-:create;comment:Create Time"`
	UpdatedAt   time.Time `gorm:"comment:Update Time"`
}

func (Group) TableName() string {
	return "subscribe_group"
}
