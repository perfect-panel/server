package application

import (
	"time"
)

type Application struct {
	Id                  int64  `gorm:"primary_key"`
	Name                string `gorm:"type:varchar(255);default:'';not null;comment:应用名称"`
	Icon                string `gorm:"type:text;not null;comment:应用图标"`
	Description         string `gorm:"type:text;comment:更新描述"`
	SubscribeType       string `gorm:"type:varchar(50);default:'';not null;comment:订阅类型"`
	ApplicationVersions []ApplicationVersion
	CreatedAt           time.Time `gorm:"<-:create;comment:创建时间"`
	UpdatedAt           time.Time `gorm:"comment:更新时间"`
}

func (Application) TableName() string {
	return "application"
}

type ApplicationVersion struct {
	Id            int64     `gorm:"primary_key"`
	Url           string    `gorm:"type:varchar(255);default:'';not null;comment:应用地址"`
	Version       string    `gorm:"type:varchar(255);default:'';not null;comment:应用版本"`
	Platform      string    `gorm:"type:varchar(50);default:'';not null;comment:应用平台"`
	IsDefault     bool      `gorm:"type:tinyint(1);not null;default:0;comment:默认版本"`
	Description   string    `gorm:"type:text;comment:更新描述"`
	ApplicationId int64     `gorm:"comment:所属应用"`
	CreatedAt     time.Time `gorm:"<-:create;comment:创建时间"`
	UpdatedAt     time.Time `gorm:"comment:更新时间"`
}

func (ApplicationVersion) TableName() string {
	return "application_version"
}

type ApplicationConfig struct {
	Id                     int64     `gorm:"primary_key"`
	AppId                  int64     `gorm:"type:int;not null;default:0;comment:App id"`
	EncryptionKey          string    `gorm:"type:text;comment:Encryption Key"`
	EncryptionMethod       string    `gorm:"type:varchar(255);comment:Encryption Method"`
	Domains                string    `gorm:"type:text;comment:Domains"`
	StartupPicture         string    `gorm:"type:text;comment:Startup Picture"`
	StartupPictureSkipTime int64     `gorm:"type:int;not null;default:0;comment:Startup Picture Skip Time"`
	InvitationLink         string    `gorm:"type:text;comment:Invitation Link"`
	KrWebsiteId            string    `gorm:"type:varchar(255);default:'';comment:Kr Website ID"`
	CreatedAt              time.Time `gorm:"<-:create;comment:Create Time"`
	UpdatedAt              time.Time `gorm:"comment:Update Time"`
}

func (ApplicationConfig) TableName() string {
	return "application_config"
}
