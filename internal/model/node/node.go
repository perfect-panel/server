package node

import (
	"time"

	"gorm.io/gorm"
)

type Node struct {
	Id        int64     `gorm:"primary_key"`
	Name      string    `gorm:"type:varchar(100);not null;default:'';comment:Node Name"`
	Tags      string    `gorm:"type:varchar(255);not null;default:'';comment:Tags"`
	Port      uint16    `gorm:"not null;default:0;comment:Connect Port"`
	Address   string    `gorm:"type:varchar(255);not null;default:'';comment:Connect Address"`
	ServerId  int64     `gorm:"not null;default:0;comment:Server ID"`
	Server    *Server   `gorm:"foreignKey:ServerId;references:Id"`
	Protocol  string    `gorm:"type:varchar(100);not null;default:'';comment:Protocol"`
	Enabled   *bool     `gorm:"type:boolean;not null;default:true;comment:Enabled"`
	Sort      int       `gorm:"uniqueIndex;not null;default:0;comment:Sort"`
	CreatedAt time.Time `gorm:"<-:create;comment:Creation Time"`
	UpdatedAt time.Time `gorm:"comment:Update Time"`
}

func (n *Node) TableName() string {
	return "nodes"
}

func (n *Node) BeforeCreate(tx *gorm.DB) error {
	if n.Sort == 0 {
		var maxSort int
		if err := tx.Model(&Node{}).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort).Error; err != nil {
			return err
		}
		n.Sort = maxSort + 1
	}
	return nil
}
