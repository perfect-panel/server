package order

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	queue "github.com/perfect-panel/server/queue/types"
)

type UpdateOrderStatusLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update order status
func NewUpdateOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrderStatusLogic {
	return &UpdateOrderStatusLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateOrderStatusLogic) UpdateOrderStatus(req *types.UpdateOrderStatusRequest) error {
	info, err := l.svcCtx.OrderModel.FindOne(l.ctx, req.Id)
	if err != nil {
		l.Errorw("[UpdateOrderStatus] FindOne error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FindOne error: %v", err.Error())
	}

	if req.PaymentId != 0 {
		paymentMethod, err := l.svcCtx.PaymentModel.FindOne(l.ctx, req.PaymentId)
		if err != nil {
			l.Logger.Error("[CreateOrder] PaymentMethod Not Found", logger.Field("error", err.Error()))
			return errors.Wrapf(xerr.NewErrCode(xerr.PaymentMethodNotFound), "PaymentMethod not found: %v", err.Error())
		}
		info.PaymentId = paymentMethod.Id
		info.Method = paymentMethod.Platform
	}
	if req.TradeNo != "" {
		info.TradeNo = req.TradeNo
	}

	err = l.svcCtx.OrderModel.Transaction(l.ctx, func(db *gorm.DB) error {
		if err := l.svcCtx.OrderModel.Update(l.ctx, info, db); err != nil {
			l.Errorw("[UpdateOrderStatus] Update error", logger.Field("error", err.Error()), logger.Field("OrderID", info.Id))
			return err
		}
		if err := l.svcCtx.OrderModel.UpdateOrderStatus(l.ctx, info.OrderNo, req.Status, db); err != nil {
			return err
		}
		// If order status is 2, create user subscription
		if req.Status == 2 {
			payload := queue.ForthwithActivateOrderPayload{
				OrderNo: info.OrderNo,
			}
			p, _ := json.Marshal(payload)
			task := asynq.NewTask(queue.ForthwithActivateOrder, p)
			_, err = l.svcCtx.Queue.EnqueueContext(l.ctx, task)
			if err != nil {
				l.Errorw("[UpdateOrderStatus] Enqueue error", logger.Field("error", err.Error()))
				return errors.Wrapf(xerr.NewErrCode(xerr.QueueEnqueueError), "Enqueue error: %v", err.Error())
			}
		}
		return nil
	})
	if err != nil {
		l.Errorw("[UpdateOrderStatus] Transaction error", logger.Field("error", err.Error()))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Transaction error: %v", err.Error())
	}
	return nil
}
