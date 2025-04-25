package payment

import (
	"context"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetAvailablePaymentMethodsLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetAvailablePaymentMethodsLogic Get available payment methods
func NewGetAvailablePaymentMethodsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAvailablePaymentMethodsLogic {
	return &GetAvailablePaymentMethodsLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAvailablePaymentMethodsLogic) GetAvailablePaymentMethods() (resp *types.GetAvailablePaymentMethodsResponse, err error) {
	data, err := l.svcCtx.PaymentModel.FindAvailableMethods(l.ctx)
	if err != nil {
		l.Error("[GetAvailablePaymentMethods] database error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "GetAvailablePaymentMethods: %v", err.Error())
	}
	resp = &types.GetAvailablePaymentMethodsResponse{
		List: make([]types.PaymentMethod, 0),
	}
	tool.DeepCopy(&resp.List, data)
	return
}
