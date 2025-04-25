package log

import (
	"gorm.io/gorm"
)

func NewModel(conn *gorm.DB) Model {
	return newLogModel(conn)
}
