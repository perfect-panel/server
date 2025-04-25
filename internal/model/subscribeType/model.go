package subscribeType

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type customSubscribeTypeLogicModel interface {
}

// NewModel returns a model for the database table.
func NewModel(conn *gorm.DB, c *redis.Client) Model {
	return &customSubscribeTypeModel{
		defaultSubscribeTypeModel: newSubscribeTypeModel(conn, c),
	}
}
