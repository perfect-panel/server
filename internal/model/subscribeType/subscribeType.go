package subscribeType

import "time"

type SubscribeType struct {
	Id        int64     `gorm:"primary_key"`
	Name      string    `gorm:"type:varchar(50);default:'';not null;comment:订阅类型"`
	Mark      string    `gorm:"type:varchar(255);default:'';not null;comment:订阅标识"`
	CreatedAt time.Time `gorm:"<-:create;comment:创建时间"`
	UpdatedAt time.Time `gorm:"comment:更新时间"`
}

func (SubscribeType) TableName() string {
	return "subscribe_type"
}
