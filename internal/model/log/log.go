package log

import "time"

type MessageType int

const (
	Email MessageType = iota + 1
	Mobile
)

func (t MessageType) String() string {
	switch t {
	case Email:
		return "email"
	case Mobile:
		return "mobile"
	}
	return "unknown"
}

type MessageLog struct {
	Id        int64     `gorm:"primaryKey"`
	Type      string    `gorm:"type:varchar(50);not null;default:'email';comment:Message Type"`
	Platform  string    `gorm:"type:varchar(50);not null;default:'smtp';comment:Platform"`
	To        string    `gorm:"type:text;not null;comment:To"`
	Subject   string    `gorm:"type:varchar(255);not null;default:'';comment:Subject"`
	Content   string    `gorm:"type:text;comment:Content"`
	Status    int       `gorm:"type:tinyint(1);not null;default:0;comment:Status"`
	CreatedAt time.Time `gorm:"<-:create;comment:Create Time"`
	UpdatedAt time.Time `gorm:"comment:Update Time"`
}

func (m *MessageLog) TableName() string {
	return "message_log"
}

type MessageLogFilterParams struct {
	Type     string
	Platform string
	To       string
	Subject  string
	Content  string
	Status   int
}
