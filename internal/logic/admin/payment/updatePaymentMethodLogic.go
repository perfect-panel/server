package payment

import (
	"context"
	"encoding/json"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/payment"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpdatePaymentMethodLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdatePaymentMethodLogic Update Payment Method
func NewUpdatePaymentMethodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePaymentMethodLogic {
	return &UpdatePaymentMethodLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePaymentMethodLogic) UpdatePaymentMethod(req *types.UpdatePaymentMethodRequest) (resp *types.PaymentConfig, err error) {
	if payment.ParsePlatform(req.Platform) == payment.UNSUPPORTED {
		l.Errorw("unsupported payment platform", logger.Field("mark", req.Platform))
		return nil, errors.Wrapf(xerr.NewErrCodeMsg(400, "UNSUPPORTED_PAYMENT_PLATFORM"), "unsupported payment platform: %s", req.Platform)
	}
	method, err := l.svcCtx.PaymentModel.FindOne(l.ctx, req.Id)
	if err != nil {
		l.Errorw("find payment method error", logger.Field("id", req.Id), logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find payment method error: %s", err.Error())
	}
	config := parsePaymentPlatformConfig(l.ctx, payment.ParsePlatform(req.Platform), req.Config)
	tool.DeepCopy(method, req, tool.CopyWithIgnoreEmpty(false))
	method.Config = config
	if err := l.svcCtx.PaymentModel.Update(l.ctx, method); err != nil {
		l.Errorw("update payment method error", logger.Field("id", req.Id), logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update payment method error: %s", err.Error())
	}
	resp = &types.PaymentConfig{}
	tool.DeepCopy(resp, method)
	var configMap map[string]interface{}
	_ = json.Unmarshal([]byte(method.Config), &configMap)
	resp.Config = configMap
	return
}
