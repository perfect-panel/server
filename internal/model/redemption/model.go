package redemption

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// NewRedemptionCodeModel returns a model for the redemption_code table.
func NewRedemptionCodeModel(conn *gorm.DB, c *redis.Client) RedemptionCodeModel {
	return &customRedemptionCodeModel{
		defaultRedemptionCodeModel: newRedemptionCodeModel(conn, c),
	}
}

// NewRedemptionRecordModel returns a model for the redemption_record table.
func NewRedemptionRecordModel(conn *gorm.DB, c *redis.Client) RedemptionRecordModel {
	return &customRedemptionRecordModel{
		defaultRedemptionRecordModel: newRedemptionRecordModel(conn, c),
	}
}
