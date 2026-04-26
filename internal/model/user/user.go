package user

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id                    int64          `gorm:"primaryKey"`
	Password              string         `gorm:"type:varchar(100);not null;comment:User Password"`
	Algo                  string         `gorm:"type:varchar(20);default:'default';comment:Encryption Algorithm"`
	Salt                  string         `gorm:"type:varchar(20);default:null;comment:Password Salt"`
	Avatar                string         `gorm:"type:MEDIUMTEXT;comment:User Avatar"`
	Balance               int64          `gorm:"default:0;comment:User Balance"` // User Balance Amount
	ReferCode             string         `gorm:"type:varchar(20);default:'';comment:Referral Code"`
	RefererId             int64          `gorm:"index:idx_referer;comment:Referrer ID"`
	Commission            int64          `gorm:"default:0;comment:Commission"`                      // Commission Amount
	ReferralPercentage    uint8          `gorm:"default:0;comment:Referral"`                        // Referral Percentage
	OnlyFirstPurchase     *bool          `gorm:"default:true;not null;comment:Only First Purchase"` // Only First Purchase Referral
	GiftAmount            int64          `gorm:"default:0;comment:User Gift Amount"`
	Enable                *bool          `gorm:"default:true;not null;comment:Is Account Enabled"`
	IsAdmin               *bool          `gorm:"default:false;not null;comment:Is Admin"`
	EnableBalanceNotify   *bool          `gorm:"default:false;not null;comment:Enable Balance Change Notifications"`
	EnableLoginNotify     *bool          `gorm:"default:false;not null;comment:Enable Login Notifications"`
	EnableSubscribeNotify *bool          `gorm:"default:false;not null;comment:Enable Subscription Notifications"`
	EnableTradeNotify     *bool          `gorm:"default:false;not null;comment:Enable Trade Notifications"`
	AuthMethods           []AuthMethods  `gorm:"foreignKey:UserId;references:Id"`
	UserDevices           []Device       `gorm:"foreignKey:UserId;references:Id"`
	Rules                 string         `gorm:"type:TEXT;comment:User Rules"`
	// 用户标签:逗号分隔的字符串,管理员自由打标(如 VIP, 试用, 长期客户)。GORM AutoMigrate 会建列。
	Tags                  string         `gorm:"column:tags;type:varchar(255);not null;default:'';comment:User Tags (comma-separated)"`
	// V4.4 #45:用户上次接受的 terms_of_use 版本字符串。空 = 从未接受。
	// 与 site_content.version 比较;不等则前端拦截让用户重新接受。
	TermsVersion          string         `gorm:"column:terms_version;type:varchar(32);not null;default:''"`
	CreatedAt             time.Time      `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt             time.Time      `gorm:"comment:Update Time"`
	DeletedAt             gorm.DeletedAt `gorm:"index;comment:Deletion Time"`
}

func (*User) TableName() string {
	return "user"
}

type Subscribe struct {
	Id           int64      `gorm:"primaryKey"`
	UserId       int64      `gorm:"index:idx_user_id;not null;comment:User ID"`
	User         User       `gorm:"foreignKey:UserId;references:Id"`
	OrderId      int64      `gorm:"index:idx_order_id;not null;comment:Order ID"`
	SubscribeId  int64      `gorm:"index:idx_subscribe_id;not null;comment:Subscription ID"`
	DeviceCount  int64      `gorm:"column:device_count;type:bigint;not null;default:1;comment:Allowed device slots - V4.3"`
	StartTime    time.Time  `gorm:"default:CURRENT_TIMESTAMP(3);not null;comment:Subscription Start Time"`
	ExpireTime   time.Time  `gorm:"default:NULL;comment:Subscription Expire Time"`
	FinishedAt   *time.Time `gorm:"default:NULL;comment:Finished Time"`
	Traffic      int64      `gorm:"default:0;comment:Traffic"`
	TrafficAddon int64      `gorm:"column:traffic_addon;type:bigint;not null;default:0;comment:Addon traffic bytes added to pool"`
	Download     int64      `gorm:"default:0;comment:Download Traffic"`
	Upload       int64      `gorm:"default:0;comment:Upload Traffic"`
	Token        string     `gorm:"index:idx_token;type:varchar(255);default:'';comment:Token (legacy, deprecated by V4.3 device tokens)"`
	UUID         string     `gorm:"type:varchar(255);index:idx_uuid;default:'';comment:UUID (legacy, deprecated by V4.3 device uuids)"`
	Status       uint8      `gorm:"type:tinyint(1);default:0;comment:Subscription Status: 0: Pending 1: Active 2: Finished 3: Expired 4: Deducted 5: stopped"`
	ThrottledAt  *time.Time `gorm:"column:throttled_at;default:NULL;comment:Throttle start time (1Mbps after over-quota)"`
	CutOffAt     *time.Time `gorm:"column:cut_off_at;default:NULL;comment:Cut-off time = throttled_at + 24h"`
	Notified90   bool       `gorm:"column:notified_90;type:tinyint(1);not null;default:0"`
	Notified100  bool       `gorm:"column:notified_100;type:tinyint(1);not null;default:0"`
	Notified12h  bool       `gorm:"column:notified_12h;type:tinyint(1);not null;default:0"`
	Notified24h  bool       `gorm:"column:notified_24h;type:tinyint(1);not null;default:0"`
	Note         string     `gorm:"type:varchar(500);default:'';comment:User note for subscription"`
	CreatedAt    time.Time  `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt    time.Time  `gorm:"comment:Update Time"`
}

// SubscribeDevice — V4.3 设备槽。每个 user_subscribe 对应 N 个 SubscribeDevice。
// 每槽独立 UUID + Token + 订阅 URL,客户端零适配。
type SubscribeDevice struct {
	Id              int64      `gorm:"primaryKey"`
	UserSubscribeId int64      `gorm:"column:user_subscribe_id;index:idx_user_subscribe_id;not null;comment:User subscribe id"`
	UserId          int64      `gorm:"column:user_id;index:idx_user_id;not null;comment:Denormalized user id"`
	DeviceName      string     `gorm:"column:device_name;type:varchar(64);not null;default:'';comment:User-friendly device name"`
	Token           string     `gorm:"column:token;type:varchar(64);unique;not null;default:'';comment:32-byte base64url token"`
	UUID            string     `gorm:"column:uuid;type:char(36);unique;not null;default:'';comment:Device UUID (proto-agnostic identity)"`
	LastSeenIP      string     `gorm:"column:last_seen_ip;type:varchar(45);not null;default:''"`
	LastSeenAt      *time.Time `gorm:"column:last_seen_at;default:NULL"`
	TodayTraffic    int64      `gorm:"column:today_traffic;not null;default:0"`
	ResetCountHour  int        `gorm:"column:reset_count_hour;type:smallint;not null;default:0"`
	ResetCountDay   int        `gorm:"column:reset_count_day;type:smallint;not null;default:0"`
	LastResetAt     *time.Time `gorm:"column:last_reset_at;default:NULL"`
	Status          uint8      `gorm:"type:tinyint(1);not null;default:1;comment:0:disabled 1:enabled"`
	// V4.3:区分套餐基础设备 vs 用户加购设备。
	// false = 套餐基础(activate 时按 plan.device_limit 建,**不可删除**)
	// true  = 用户加购(addSubscribeDevice 时建,**用户可主动删除**)
	IsAddon   bool      `gorm:"column:is_addon;type:tinyint(1);not null;default:0;comment:1=user-purchased addon, 0=plan-included base device"`
	CreatedAt time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt time.Time `gorm:"comment:Update Time"`
}

func (*SubscribeDevice) TableName() string {
	return "user_subscribe_device"
}

func (*Subscribe) TableName() string {
	return "user_subscribe"
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

type Withdrawal struct {
	Id        int64     `gorm:"primaryKey"`
	UserId    int64     `gorm:"index:idx_user_id;not null;comment:User ID"`
	Amount    int64     `gorm:"not null;comment:Withdrawal Amount"`
	Content   string    `gorm:"type:text;comment:Withdrawal Content"`
	Status    uint8     `gorm:"type:tinyint(1);default:0;comment:Withdrawal Status: 0: Pending 1: Approved 2: Rejected"`
	Reason    string    `gorm:"type:varchar(500);default:'';comment:Rejection Reason"`
	CreatedAt time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt time.Time `gorm:"comment:Update Time"`
}

func (*Withdrawal) TableName() string {
	return "user_withdrawal"
}
