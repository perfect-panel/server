package orderLogic

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/ppanel-server/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/ppanel-server/internal/logic/public/order"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	internal "github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/queue/types"
)

type DeferCloseOrderLogic struct {
	svc *svc.ServiceContext
}

func NewDeferCloseOrderLogic(svc *svc.ServiceContext) *DeferCloseOrderLogic {
	return &DeferCloseOrderLogic{
		svc: svc,
	}
}

func (l *DeferCloseOrderLogic) ProcessTask(ctx context.Context, task *asynq.Task) error {
	payload := types.DeferCloseOrderPayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logger.WithContext(ctx).Error("[DeferCloseOrderLogic] Unmarshal payload failed",
			logger.Field("error", err.Error()),
			logger.Field("payload", string(task.Payload())),
		)
		return nil
	}

	err := order.NewCloseOrderLogic(ctx, l.svc).CloseOrder(&internal.CloseOrderRequest{
		OrderNo: payload.OrderNo,
	})
	count, ok := asynq.GetRetryCount(ctx)
	if !ok {
		return nil
	}
	if err != nil && count < 3 {
		return err
	}
	return nil
}
