package user

import (
	"time"
)

type User struct {
	Id                    int64         `gorm:"primaryKey"`
	Password              string        `gorm:"type:varchar(100);not null;comment:User Password"`
	Avatar                string        `gorm:"type:MEDIUMTEXT;comment:User Avatar"`
	Balance               int64         `gorm:"default:0;comment:User Balance"` // User Balance Amount
	ReferCode             string        `gorm:"type:varchar(20);default:'';comment:Referral Code"`
	RefererId             int64         `gorm:"index:idx_referer;comment:Referrer ID"`
	Commission            int64         `gorm:"default:0;comment:Commission"` // Commission Amount
	GiftAmount            int64         `gorm:"default:0;comment:User Gift Amount"`
	Enable                *bool         `gorm:"default:true;not null;comment:Is Account Enabled"`
	IsAdmin               *bool         `gorm:"default:false;not null;comment:Is Admin"`
	EnableBalanceNotify   *bool         `gorm:"default:false;not null;comment:Enable Balance Change Notifications"`
	EnableLoginNotify     *bool         `gorm:"default:false;not null;comment:Enable Login Notifications"`
	EnableSubscribeNotify *bool         `gorm:"default:false;not null;comment:Enable Subscription Notifications"`
	EnableTradeNotify     *bool         `gorm:"default:false;not null;comment:Enable Trade Notifications"`
	AuthMethods           []AuthMethods `gorm:"foreignKey:UserId;references:Id"`
	UserDevices           []Device      `gorm:"foreignKey:UserId;references:Id"`
	CreatedAt             time.Time     `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt             time.Time     `gorm:"comment:Update Time"`
}

func (*User) TableName() string {
	return "user"
}

type Subscribe struct {
	Id          int64      `gorm:"primaryKey"`
	UserId      int64      `gorm:"index:idx_user_id;not null;comment:User ID"`
	User        User       `gorm:"foreignKey:UserId;references:Id"`
	OrderId     int64      `gorm:"index:idx_order_id;not null;comment:Order ID"`
	SubscribeId int64      `gorm:"index:idx_subscribe_id;not null;comment:Subscription ID"`
	StartTime   time.Time  `gorm:"default:CURRENT_TIMESTAMP(3);not null;comment:Subscription Start Time"`
	ExpireTime  time.Time  `gorm:"default:NULL;comment:Subscription Expire Time"`
	FinishedAt  *time.Time `gorm:"default:NULL;comment:Finished Time"`
	Traffic     int64      `gorm:"default:0;comment:Traffic"`
	Download    int64      `gorm:"default:0;comment:Download Traffic"`
	Upload      int64      `gorm:"default:0;comment:Upload Traffic"`
	Token       string     `gorm:"index:idx_token;unique;type:varchar(255);default:'';comment:Token"`
	UUID        string     `gorm:"type:varchar(255);unique;index:idx_uuid;default:'';comment:UUID"`
	Status      uint8      `gorm:"type:tinyint(1);default:0;comment:Subscription Status: 0: Pending 1: Active 2: Finished 3: Expired 4: Deducted"`
	CreatedAt   time.Time  `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt   time.Time  `gorm:"comment:Update Time"`
}

func (*Subscribe) TableName() string {
	return "user_subscribe"
}

type BalanceLog struct {
	Id        int64     `gorm:"primaryKey"`
	UserId    int64     `gorm:"index:idx_user_id;not null;comment:User ID"`
	Amount    int64     `gorm:"not null;comment:Amount"`
	Type      uint8     `gorm:"type:tinyint(1);not null;comment:Type: 1: Recharge 2: Withdraw 3: Payment 4: Refund 5: Reward"`
	OrderId   int64     `gorm:"default:null;comment:Order ID"`
	Balance   int64     `gorm:"not null;comment:Balance"`
	CreatedAt time.Time `gorm:"<-:create;comment:Creation Time"`
}

func (BalanceLog) TableName() string {
	return "user_balance_log"
}

type GiftAmountLog struct {
	Id              int64     `gorm:"primaryKey"`
	UserId          int64     `gorm:"index:idx_user_id;not null;comment:User ID"`
	UserSubscribeId int64     `gorm:"default:null;comment:Deduction User Subscribe ID"`
	OrderNo         string    `gorm:"default:null;comment:Order No."`
	Type            uint8     `gorm:"type:tinyint(1);not null;comment:Type: 1: Increase 2: Reduce"`
	Amount          int64     `gorm:"not null;comment:Amount"`
	Balance         int64     `gorm:"not null;comment:Balance"`
	Remark          string    `gorm:"type:varchar(255);default:'';comment:Remark"`
	CreatedAt       time.Time `gorm:"<-:create;comment:Creation Time"`
}

func (GiftAmountLog) TableName() string {
	return "user_gift_amount_log"
}

type CommissionLog struct {
	Id        int64     `gorm:"primaryKey"`
	UserId    int64     `gorm:"index:idx_user_id;not null;comment:User ID"`
	OrderNo   string    `gorm:"default:null;comment:Order No."`
	Amount    int64     `gorm:"not null;comment:Amount"`
	CreatedAt time.Time `gorm:"<-:create;comment:Creation Time"`
}

func (*CommissionLog) TableName() string {
	return "user_commission_log"
}

type AuthMethods struct {
	Id             int64     `gorm:"primaryKey"`
	UserId         int64     `gorm:"index:idx_user_id;not null;comment:User ID"`
	AuthType       string    `gorm:"type:varchar(255);not null;comment:Auth Type 1: apple 2: google 3: github 4: facebook 5: telegram 6: email 7: mobile 8: device"`
	AuthIdentifier string    `gorm:"type:varchar(255);unique;index:idx_auth_identifier;not null;comment:Auth Identifier"`
	Verified       bool      `gorm:"default:false;not null;comment:Is Verified"`
	CreatedAt      time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt      time.Time `gorm:"comment:Update Time"`
}

func (*AuthMethods) TableName() string {
	return "user_auth_methods"
}

type Device struct {
	Id         int64     `gorm:"primaryKey"`
	Ip         string    `gorm:"type:varchar(255);not null;comment:Device IP"`
	UserId     int64     `gorm:"index:idx_user_id;not null;comment:User ID"`
	UserAgent  string    `gorm:"default:null;comment:UserAgent."`
	Identifier string    `gorm:"type:varchar(255);unique;index:idx_identifier;default:'';comment:Device Identifier"`
	Online     bool      `gorm:"default:false;not null;comment:Online"`
	Enabled    bool      `gorm:"default:true;not null;comment:Enabled"`
	CreatedAt  time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt  time.Time `gorm:"comment:Update Time"`
}

func (*Device) TableName() string {
	return "user_device"
}

type DeviceOnlineRecord struct {
	Id            int64     `gorm:"primaryKey"`
	UserId        int64     `gorm:"type:bigint;not null;comment:User ID"`
	Identifier    string    `gorm:"type:varchar(255);not null;comment:Device Identifier"`
	OnlineTime    time.Time `gorm:"comment:Online Time"` // The time when the device goes online
	OfflineTime   time.Time `gorm:"comment:Offline Time"`
	OnlineSeconds int64     `gorm:"comment:Offline Seconds"`
	DurationDays  int64     `gorm:"comment:Duration Days"`
	CreatedAt     time.Time `gorm:"<-:create;comment:Creation Time"`
}

func (DeviceOnlineRecord) TableName() string {
	return "user_device_online_record"
}

type LoginLog struct {
	Id        int64     `gorm:"primaryKey"`
	UserId    int64     `gorm:"index:idx_user_id;not null;comment:User ID"`
	LoginIP   string    `gorm:"type:varchar(255);not null;comment:Login IP"`
	UserAgent string    `gorm:"type:text;not null;comment:UserAgent"`
	Success   *bool     `gorm:"default:false;not null;comment:Login Success"`
	CreatedAt time.Time `gorm:"<-:create;comment:Creation Time"`
}

func (LoginLog) TableName() string {
	return "user_login_log"
}

type SubscribeLog struct {
	Id              int64     `gorm:"primaryKey"`
	UserId          int64     `gorm:"index:idx_user_id;not null;comment:User ID"`
	UserSubscribeId int64     `gorm:"index:idx_user_subscribe_id;not null;comment:User Subscribe ID"`
	Token           string    `gorm:"type:varchar(255);not null;comment:Token"`
	IP              string    `gorm:"type:varchar(255);not null;comment:IP"`
	UserAgent       string    `gorm:"type:text;not null;comment:UserAgent"`
	CreatedAt       time.Time `gorm:"<-:create;comment:Creation Time"`
}

func (SubscribeLog) TableName() string {
	return "user_subscribe_log"
}

const (
	ResetSubscribeTypeAuto    uint8 = 1
	ResetSubscribeTypeAdvance uint8 = 2
	ResetSubscribeTypePaid    uint8 = 3
)

type FilterResetSubscribeLogParams struct {
	Page            int
	Size            int
	Type            uint8
	UserId          int64
	OrderNo         string
	UserSubscribeId int64
}

type ResetSubscribeLog struct {
	Id              int64     `gorm:"primaryKey"`
	UserId          int64     `gorm:"type:bigint;index:idx_user_id;not null;comment:User ID"`
	Type            uint8     `gorm:"type:tinyint(1);not null;comment:Type: 1: Auto 2: Advance 3: Paid"`
	OrderNo         string    `gorm:"type:varchar(255);default:null;comment:Order No."`
	UserSubscribeId int64     `gorm:"type:bigint;index:idx_user_subscribe_id;not null;comment:User Subscribe ID"`
	CreatedAt       time.Time `gorm:"<-:create;comment:Creation Time"`
}

func (ResetSubscribeLog) TableName() string {
	return "user_reset_subscribe_log"
}
