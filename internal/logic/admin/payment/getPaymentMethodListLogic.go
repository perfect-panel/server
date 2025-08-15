package payment

import (
	"context"
	"encoding/json"

	paymentPlatform "github.com/perfect-panel/server/pkg/payment"

	"github.com/perfect-panel/server/internal/model/payment"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetPaymentMethodListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetPaymentMethodListLogic Get Payment Method List
func NewGetPaymentMethodListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentMethodListLogic {
	return &GetPaymentMethodListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPaymentMethodListLogic) GetPaymentMethodList(req *types.GetPaymentMethodListRequest) (resp *types.GetPaymentMethodListResponse, err error) {
	total, list, err := l.svcCtx.PaymentModel.FindListByPage(l.ctx, req.Page, req.Size, &payment.Filter{
		Search: req.Search,
		Mark:   req.Platform,
		Enable: req.Enable,
	})
	if err != nil {
		l.Errorw("find payment method list error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find payment method list error: %s", err.Error())
	}
	resp = &types.GetPaymentMethodListResponse{
		Total: total,
		List:  make([]types.PaymentMethodDetail, len(list)),
	}
	for i, v := range list {
		config := make(map[string]interface{})
		_ = json.Unmarshal([]byte(v.Config), &config)
		notifyUrl := ""
		if paymentPlatform.ParsePlatform(v.Platform) != paymentPlatform.Balance {
			if v.Domain != "" {
				notifyUrl = v.Domain + "/v1/notify/" + v.Platform + "/" + v.Token
			} else {
				notifyUrl = "https://" + l.svcCtx.Config.Host + "/v1/notify/" + v.Platform + "/" + v.Token
			}
		}
		resp.List[i] = types.PaymentMethodDetail{
			Id:          v.Id,
			Name:        v.Name,
			Platform:    v.Platform,
			Icon:        v.Icon,
			Domain:      v.Domain,
			Config:      config,
			FeeMode:     v.FeeMode,
			FeePercent:  v.FeePercent,
			FeeAmount:   v.FeeAmount,
			Enable:      *v.Enable,
			NotifyURL:   notifyUrl,
			Description: v.Description,
		}
	}
	return
}
