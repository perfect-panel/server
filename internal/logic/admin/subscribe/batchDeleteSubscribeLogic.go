package subscribe

import (
	"context"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type BatchDeleteSubscribeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Batch delete subscribe
func NewBatchDeleteSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchDeleteSubscribeLogic {
	return &BatchDeleteSubscribeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

var errorIsExistActiveUser = errors.New("subscription ID belongs to an active user subscription")

func (l *BatchDeleteSubscribeLogic) BatchDeleteSubscribe(req *types.BatchDeleteSubscribeRequest) error {
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range req.Ids {
			var count int64
			// Validate whether the subscription ID belongs to an active user subscription.
			if err := tx.Model(&user.Subscribe{}).Where("subscribe_id = ? AND status = 1", id).Count(&count).Find(&user.Subscribe{}).Error; err != nil {
				l.Logger.Error("[BatchDeleteSubscribe] Query Subscribe Error: ", logger.Field("error", err.Error()))
				return err
			}
			if count > 0 {
				return errorIsExistActiveUser
			}
			if err := l.svcCtx.SubscribeModel.Delete(l.ctx, id, tx); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errorIsExistActiveUser) {
			return errors.Wrapf(xerr.NewErrCode(xerr.SubscribeIsUsedError), "subscription ID belongs to an active user subscription")
		}
		l.Logger.Error("[BatchDeleteSubscribe] Transaction Error: ", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseDeletedError), "delete subscribe failed: %v", err.Error())
	}
	return nil
}
