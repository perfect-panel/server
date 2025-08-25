package node

import "time"

type Node struct {
	Id        int64     `gorm:"primary_key"`
	Name      string    `gorm:"type:varchar(100);not null;default:'';comment:Node Name"`
	Tags      string    `gorm:"type:varchar(255);not null;default:'';comment:Tags"`
	Port      uint16    `gorm:"not null;default:0;comment:Connect Port"`
	Address   string    `gorm:"type:varchar(255);not null;default:'';comment:Connect Address"`
	ServerId  int64     `gorm:"not null;default:0;comment:Server ID"`
	Protocol  string    `gorm:"type:varchar(100);not null;default:'';comment:Protocol"`
	Enabled   *bool     `gorm:"type:boolean;not null;default:true;comment:Enabled"`
	CreatedAt time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt time.Time `gorm:"comment:Update Time"`
}

func (Node) TableName() string {
	return "nodes"
}
