package payment

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perfect-panel/server/pkg/payment/stripe"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/pkg/random"

	paymentModel "github.com/perfect-panel/server/internal/model/payment"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/payment"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type CreatePaymentMethodLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreatePaymentMethodLogic Create Payment Method
func NewCreatePaymentMethodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentMethodLogic {
	return &CreatePaymentMethodLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePaymentMethodLogic) CreatePaymentMethod(req *types.CreatePaymentMethodRequest) (resp *types.PaymentConfig, err error) {
	if payment.ParsePlatform(req.Platform) == payment.UNSUPPORTED {
		l.Errorw("unsupported payment platform", logger.Field("mark", req.Platform))
		return nil, errors.Wrapf(xerr.NewErrCodeMsg(400, "UNSUPPORTED_PAYMENT_PLATFORM"), "unsupported payment platform: %s", req.Platform)
	}
	config := parsePaymentPlatformConfig(l.ctx, payment.ParsePlatform(req.Platform), req.Config)
	var paymentMethod = &paymentModel.Payment{
		Name:        req.Name,
		Platform:    req.Platform,
		Icon:        req.Icon,
		Domain:      req.Domain,
		Description: req.Description,
		Config:      config,
		FeeMode:     req.FeeMode,
		FeePercent:  req.FeePercent,
		FeeAmount:   req.FeeAmount,
		Enable:      req.Enable,
		Token:       random.KeyNew(8, 1),
	}
	err = l.svcCtx.PaymentModel.Transaction(l.ctx, func(tx *gorm.DB) error {

		if req.Platform == "Stripe" {
			var cfg paymentModel.StripeConfig
			if err := cfg.Unmarshal(paymentMethod.Config); err != nil {
				l.Errorf("[CreatePaymentMethod] unmarshal stripe config error: %s", err.Error())
				return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "unmarshal stripe config error: %s", err.Error())
			}
			if cfg.SecretKey == "" {
				l.Error("[CreatePaymentMethod] stripe secret key is empty")
				return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "stripe secret key is empty")
			}

			// Create Stripe webhook endpoint
			client := stripe.NewClient(stripe.Config{
				SecretKey: cfg.SecretKey,
				PublicKey: cfg.PublicKey,
			})
			url := fmt.Sprintf("%s/notify/Stripe/%s", req.Domain, paymentMethod.Token)
			endpoint, err := client.CreateWebhookEndpoint(url)
			if err != nil {
				l.Errorw("[CreatePaymentMethod] create stripe webhook endpoint error", logger.Field("error", err.Error()))
				return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "create stripe webhook endpoint error: %s", err.Error())
			}
			cfg.WebhookSecret = endpoint.Secret
			paymentMethod.Config = cfg.Marshal()
		}
		if err = tx.Model(&paymentModel.Payment{}).Create(paymentMethod).Error; err != nil {
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "insert payment method error: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	resp = &types.PaymentConfig{}
	tool.DeepCopy(resp, paymentMethod)
	var configMap map[string]interface{}
	_ = json.Unmarshal([]byte(paymentMethod.Config), &configMap)
	resp.Config = configMap
	return
}

func parsePaymentPlatformConfig(ctx context.Context, platform payment.Platform, config interface{}) string {
	data, err := json.Marshal(config)
	if err != nil {
		logger.WithContext(ctx).Errorw("parse payment platform config error", logger.Field("platform", platform), logger.Field("config", config), logger.Field("error", err.Error()))
	}
	switch platform {
	case payment.Stripe:
		stripe := &paymentModel.StripeConfig{}
		if err := stripe.Unmarshal(string(data)); err != nil {
			logger.WithContext(ctx).Errorw("parse stripe config error", logger.Field("config", string(data)), logger.Field("error", err.Error()))
		}
		return stripe.Marshal()
	case payment.AlipayF2F:
		alipay := &paymentModel.AlipayF2FConfig{}
		if err := alipay.Unmarshal(string(data)); err != nil {
			logger.WithContext(ctx).Errorw("parse alipay config error", logger.Field("config", string(data)), logger.Field("error", err.Error()))
		}
		return alipay.Marshal()
	case payment.EPay:
		epay := &paymentModel.EPayConfig{}
		if err := epay.Unmarshal(string(data)); err != nil {
			logger.WithContext(ctx).Errorw("parse epay config error", logger.Field("config", string(data)), logger.Field("error", err.Error()))
		}
		return epay.Marshal()
	default:
		return ""
	}
}
