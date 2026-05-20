package user

import (
	"context"

	"gorm.io/gorm"
)

func (m *customUserModel) InsertWithdrawal(ctx context.Context, data *Withdrawal, tx ...*gorm.DB) error {
	return m.ExecNoCacheCtx(ctx, func(conn *gorm.DB) error {
		if len(tx) > 0 {
			conn = tx[0]
		}
		return conn.Create(data).Error
	})
}
