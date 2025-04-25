package application

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type customApplicationLogicModel interface {
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customApplicationModel{
		defaultApplicationModel: newApplicationModel(conn, c),
	}
}
