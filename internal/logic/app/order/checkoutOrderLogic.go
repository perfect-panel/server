package order

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	paymentPlatform "github.com/perfect-panel/ppanel-server/pkg/payment"

	"github.com/perfect-panel/ppanel-server/pkg/constant"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/ppanel-server/internal/model/order"
	"github.com/perfect-panel/ppanel-server/internal/model/payment"
	"github.com/perfect-panel/ppanel-server/internal/model/user"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/exchangeRate"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/payment/alipay"
	"github.com/perfect-panel/ppanel-server/pkg/payment/epay"
	"github.com/perfect-panel/ppanel-server/pkg/payment/stripe"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	"github.com/perfect-panel/ppanel-server/pkg/xerr"
	queueType "github.com/perfect-panel/ppanel-server/queue/types"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CheckoutOrderLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type CurrencyConfig struct {
	CurrencyUnit   string
	CurrencySymbol string
	AccessKey      string
}

const (
	Stripe = "Stripe"
	QR     = "qr"
	Link   = "link"
)

// NewCheckoutOrderLogic Checkout order
func NewCheckoutOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckoutOrderLogic {
	return &CheckoutOrderLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckoutOrderLogic) CheckoutOrder(req *types.CheckoutOrderRequest, requestHost string) (resp *types.CheckoutOrderResponse, err error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		l.Error("[CheckoutOrderLogic] Invalid access")
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid access")
	}
	// find order
	orderInfo, err := l.svcCtx.OrderModel.FindOneByOrderNo(l.ctx, req.OrderNo)
	if err != nil {
		l.Error("[CheckoutOrderLogic] FindOneByOrderNo error", logger.Field("orderNo", req.OrderNo), logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FindOneByOrderNo error: %s", err.Error())
	}

	if orderInfo.Status != 1 {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Order status error")
	}

	paymentConfig, err := l.svcCtx.PaymentModel.FindOne(l.ctx, orderInfo.PaymentId)
	if err != nil {
		l.Error("[CheckoutOrderLogic] FindOneByPaymentMark error", logger.Field("paymentMark", orderInfo.Method), logger.Field("PaymentID", orderInfo.PaymentId), logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FindOneByPaymentMark error: %s", err.Error())
	}
	var stripePayment *types.StripePayment = nil
	var url, t string

	// switch payment method
	switch paymentPlatform.ParsePlatform(paymentConfig.Platform) {
	case paymentPlatform.Stripe:
		result, err := l.stripePayment(paymentConfig.Config, orderInfo, u)
		if err != nil {
			l.Error("[CheckoutOrderLogic] stripePayment error", logger.Field("error", err.Error()))
			return nil, err
		}
		stripePayment = result
		t = Stripe
	case paymentPlatform.EPay:
		// epay
		url, err = l.epayPayment(paymentConfig, orderInfo, req.ReturnUrl, requestHost)
		if err != nil {
			l.Error("[CheckoutOrderLogic] epayPayment error", logger.Field("error", err.Error()))
			return nil, err
		}
		t = Link
	case paymentPlatform.AlipayF2F:
		// alipay f2f
		url, err = l.alipayF2fPayment(paymentConfig, orderInfo, requestHost)
		if err != nil {
			return nil, err
		}
		t = QR
	case paymentPlatform.Balance:
		// balance
		if err = l.balancePayment(u, orderInfo); err != nil {
			return nil, err
		}
		t = paymentPlatform.Balance.String()
	default:
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Payment method not supported")
	}
	return &types.CheckoutOrderResponse{
		Type:        t,
		CheckoutUrl: url,
		Stripe:      stripePayment,
	}, nil
}

// Query exchange rate
func (l *CheckoutOrderLogic) queryExchangeRate(to string, src int64) (amount float64, err error) {
	amount = float64(src) / float64(100)
	// query system currency
	currency, err := l.svcCtx.SystemModel.GetCurrencyConfig(l.ctx)
	if err != nil {
		l.Error("[CheckoutOrderLogic] GetCurrencyConfig error", logger.Field("error", err.Error()))
		return 0, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "GetCurrencyConfig error: %s", err.Error())
	}
	configs := &CurrencyConfig{}
	tool.SystemConfigSliceReflectToStruct(currency, configs)
	if configs.AccessKey == "" {
		return amount, nil
	}
	if configs.CurrencyUnit != to {
		// query exchange rate
		result, err := exchangeRate.GetExchangeRete(configs.CurrencyUnit, to, configs.AccessKey, 1)
		if err != nil {
			return 0, err
		}
		amount = result * amount
	}
	return amount, nil
}

// Stripe Payment
func (l *CheckoutOrderLogic) stripePayment(config string, info *order.Order, u *user.User) (*types.StripePayment, error) {
	// stripe WeChat pay or stripe alipay
	stripeConfig := payment.StripeConfig{}
	if err := json.Unmarshal([]byte(config), &stripeConfig); err != nil {
		l.Error("[CheckoutOrderLogic] Unmarshal error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Unmarshal error: %s", err.Error())
	}
	client := stripe.NewClient(stripe.Config{
		SecretKey:     stripeConfig.SecretKey,
		PublicKey:     stripeConfig.PublicKey,
		WebhookSecret: stripeConfig.WebhookSecret,
	})
	// Calculate the amount with exchange rate
	amount, err := l.queryExchangeRate("CNY", info.Amount)
	if err != nil {
		l.Error("[CheckoutOrderLogic] queryExchangeRate error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "queryExchangeRate error: %s", err.Error())
	}
	convertAmount := int64(amount * 100)
	// create payment
	result, err := client.CreatePaymentSheet(&stripe.Order{
		OrderNo:   info.OrderNo,
		Subscribe: strconv.FormatInt(info.SubscribeId, 10),
		Amount:    convertAmount,
		Currency:  "cny",
		Payment:   stripeConfig.Payment,
	},
		&stripe.User{
			UserId: u.Id,
		})
	if err != nil {
		l.Error("[CheckoutOrderLogic] CreatePaymentSheet error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "CreatePaymentSheet error: %s", err.Error())
	}
	tradeNo := result.TradeNo
	stripePayment := &types.StripePayment{
		PublishableKey: stripeConfig.PublicKey,
		ClientSecret:   result.ClientSecret,
		Method:         stripeConfig.Payment,
	}
	// save payment
	info.TradeNo = tradeNo
	err = l.svcCtx.OrderModel.Update(l.ctx, info)
	if err != nil {
		l.Error("[CheckoutOrderLogic] Update error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Update error: %s", err.Error())
	}
	return stripePayment, nil
}

// epay payment
func (l *CheckoutOrderLogic) epayPayment(config *payment.Payment, info *order.Order, returnUrl, requestHost string) (string, error) {
	epayConfig := payment.EPayConfig{}
	if err := json.Unmarshal([]byte(config.Config), &epayConfig); err != nil {
		l.Error("[CheckoutOrderLogic] Unmarshal error", logger.Field("error", err.Error()))
		return "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Unmarshal error: %s", err.Error())
	}
	client := epay.NewClient(epayConfig.Pid, epayConfig.Url, epayConfig.Key)
	// Calculate the amount with exchange rate
	amount, err := l.queryExchangeRate("CNY", info.Amount)
	if err != nil {
		return "", err
	}
	var domain string
	if config.Domain != "" {
		domain = config.Domain
	} else {
		domain = fmt.Sprintf("http://%s", requestHost)
	}
	// create payment
	url := client.CreatePayUrl(epay.Order{
		Name:      l.svcCtx.Config.Site.SiteName,
		Amount:    amount,
		OrderNo:   info.OrderNo,
		SignType:  "MD5",
		NotifyUrl: domain + "/v1/notify/epay",
		ReturnUrl: returnUrl,
	})
	return url, nil
}

// alipay f2f payment
func (l *CheckoutOrderLogic) alipayF2fPayment(pay *payment.Payment, info *order.Order, requestHost string) (string, error) {
	f2FConfig := payment.AlipayF2FConfig{}
	if err := json.Unmarshal([]byte(pay.Config), &f2FConfig); err != nil {
		l.Error("[CheckoutOrderLogic] Unmarshal error", logger.Field("error", err.Error()))
		return "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Unmarshal error: %s", err.Error())
	}
	var domain string
	if pay.Domain != "" {
		domain = pay.Domain
	} else {
		domain = fmt.Sprintf("http://%s", requestHost)
	}
	client := alipay.NewClient(alipay.Config{
		AppId:       f2FConfig.AppId,
		PrivateKey:  f2FConfig.PrivateKey,
		PublicKey:   f2FConfig.PublicKey,
		InvoiceName: f2FConfig.InvoiceName,
		NotifyURL:   domain + "/notify/alipay",
	})
	// Calculate the amount with exchange rate
	amount, err := l.queryExchangeRate("CNY", info.Amount)
	if err != nil {
		l.Error("[CheckoutOrderLogic] queryExchangeRate error", logger.Field("error", err.Error()))
		return "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "queryExchangeRate error: %s", err.Error())
	}
	convertAmount := int64(amount * 100)
	// create payment
	QRCode, err := client.PreCreateTrade(l.ctx, alipay.Order{
		OrderNo: info.OrderNo,
		Amount:  convertAmount,
	})
	if err != nil {
		l.Error("[CheckoutOrderLogic] PreCreateTrade error", logger.Field("error", err.Error()))
		return "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "PreCreateTrade error: %s", err.Error())
	}
	return QRCode, nil
}

// Balance payment
func (l *CheckoutOrderLogic) balancePayment(u *user.User, o *order.Order) error {
	var userInfo user.User
	err := l.svcCtx.UserModel.Transaction(l.ctx, func(db *gorm.DB) error {
		err := db.Model(&user.User{}).Where("id = ?", u.Id).First(&userInfo).Error
		if err != nil {
			return err
		}

		if userInfo.Balance < o.Amount {
			return errors.Wrapf(xerr.NewErrCode(xerr.InsufficientBalance), "Insufficient balance")
		}
		// deduct balance
		userInfo.Balance -= o.Amount
		err = l.svcCtx.UserModel.Update(l.ctx, &userInfo)
		if err != nil {
			return err
		}
		// create balance log
		balanceLog := &user.BalanceLog{
			Id:      0,
			UserId:  u.Id,
			Amount:  o.Amount,
			Type:    3,
			OrderId: o.Id,
			Balance: userInfo.Balance,
		}
		err = db.Create(balanceLog).Error
		if err != nil {
			return err
		}
		return l.svcCtx.OrderModel.UpdateOrderStatus(l.ctx, o.OrderNo, 2)
	})
	if err != nil {
		l.Error("[CheckoutOrderLogic] Transaction error", logger.Field("error", err.Error()), logger.Field("orderNo", o.OrderNo))
		return err
	}
	// create activity order task
	payload := queueType.ForthwithActivateOrderPayload{
		OrderNo: o.OrderNo,
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		l.Error("[CheckoutOrderLogic] Marshal error", logger.Field("error", err.Error()))
		return err
	}

	task := asynq.NewTask(queueType.ForthwithActivateOrder, bytes)
	_, err = l.svcCtx.Queue.EnqueueContext(l.ctx, task)
	if err != nil {
		l.Error("[CheckoutOrderLogic] Enqueue error", logger.Field("error", err.Error()))
		return err
	}
	l.Logger.Info("[CheckoutOrderLogic] Enqueue success", logger.Field("orderNo", o.OrderNo))
	return nil
}
