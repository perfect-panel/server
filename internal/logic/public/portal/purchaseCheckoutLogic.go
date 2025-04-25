package portal

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/perfect-panel/server/pkg/constant"

	paymentPlatform "github.com/perfect-panel/server/pkg/payment"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/user"
	queueType "github.com/perfect-panel/server/queue/types"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/payment"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/exchangeRate"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/payment/alipay"
	"github.com/perfect-panel/server/pkg/payment/epay"
	"github.com/perfect-panel/server/pkg/payment/stripe"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type PurchaseCheckoutLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPurchaseCheckoutLogic Purchase Checkout
func NewPurchaseCheckoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PurchaseCheckoutLogic {
	return &PurchaseCheckoutLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PurchaseCheckoutLogic) PurchaseCheckout(req *types.CheckoutOrderRequest) (resp *types.CheckoutOrderResponse, err error) {
	// Find order
	orderInfo, err := l.svcCtx.OrderModel.FindOneByOrderNo(l.ctx, req.OrderNo)
	if err != nil {
		l.Logger.Error("[PurchaseCheckout] Find order failed", logger.Field("error", err.Error()), logger.Field("orderNo", req.OrderNo))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.OrderNotExist), "order not exist: %v", req.OrderNo)
	}

	if orderInfo.Status != 1 {
		l.Logger.Error("[PurchaseCheckout] Order status error", logger.Field("status", orderInfo.Status))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.OrderStatusError), "order status error: %v", orderInfo.Status)
	}

	// find payment method
	paymentConfig, err := l.svcCtx.PaymentModel.FindOne(l.ctx, orderInfo.PaymentId)
	if err != nil {
		l.Logger.Error("[Purchase] Database query error", logger.Field("error", err.Error()), logger.Field("payment", orderInfo.Method))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find payment method error: %v", err.Error())
	}
	switch paymentPlatform.ParsePlatform(orderInfo.Method) {
	case paymentPlatform.EPay:
		url, err := l.epayPayment(paymentConfig, orderInfo, req.ReturnUrl)
		if err != nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "epayPayment error: %v", err.Error())
		}
		resp = &types.CheckoutOrderResponse{
			CheckoutUrl: url,
			Type:        "url",
		}
	case paymentPlatform.Stripe:
		stripePayment, err := l.stripePayment(paymentConfig.Config, orderInfo, "")
		if err != nil {
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "stripePayment error: %v", err.Error())
		}
		resp = &types.CheckoutOrderResponse{
			Type:   "stripe",
			Stripe: stripePayment,
		}
	case paymentPlatform.AlipayF2F:
		url, err := l.alipayF2fPayment(paymentConfig, orderInfo)
		if err != nil {
			l.Errorw("[CheckoutOrderLogic] alipayF2fPayment error", logger.Field("error", err.Error()))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "alipayF2fPayment error: %v", err.Error())
		}
		resp = &types.CheckoutOrderResponse{
			Type:        "qr",
			CheckoutUrl: url,
		}
	case paymentPlatform.Balance:
		if orderInfo.UserId == 0 {
			l.Errorw("[CheckoutOrderLogic] user not found", logger.Field("userId", orderInfo.UserId))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.UserNotExist), "user not found")
		}
		// find user
		userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, orderInfo.UserId)
		if err != nil {
			l.Errorw("[CheckoutOrderLogic] FindOne User error", logger.Field("error", err.Error()))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "FindOne error: %s", err.Error())
		}

		// balance
		if err = l.balancePayment(userInfo, orderInfo); err != nil {
			return nil, err
		}
		resp = &types.CheckoutOrderResponse{
			Type: "balance",
		}

	default:
		l.Errorw("[CheckoutOrderLogic] payment method not found", logger.Field("method", orderInfo.Method))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "payment method not found")
	}
	return
}

// alipay f2f payment
func (l *PurchaseCheckoutLogic) alipayF2fPayment(pay *payment.Payment, info *order.Order) (string, error) {
	f2FConfig := payment.AlipayF2FConfig{}
	if err := json.Unmarshal([]byte(pay.Config), &f2FConfig); err != nil {
		l.Errorw("[PurchaseCheckoutLogic] Unmarshal error", logger.Field("error", err.Error()))
		return "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Unmarshal error: %s", err.Error())
	}
	notifyUrl := ""
	if pay.Domain != "" {
		notifyUrl = pay.Domain + "/v1/notify/" + pay.Platform + "/" + pay.Token
	} else {
		host, ok := l.ctx.Value(constant.CtxKeyRequestHost).(string)
		if !ok {
			host = l.svcCtx.Config.Host
		}
		notifyUrl = "https://" + host + "/v1/notify/" + pay.Platform + "/" + pay.Token
	}
	client := alipay.NewClient(alipay.Config{
		AppId:       f2FConfig.AppId,
		PrivateKey:  f2FConfig.PrivateKey,
		PublicKey:   f2FConfig.PublicKey,
		InvoiceName: f2FConfig.InvoiceName,
		NotifyURL:   notifyUrl,
	})
	// Calculate the amount with exchange rate
	amount, err := l.queryExchangeRate("CNY", info.Amount)
	if err != nil {
		l.Errorw("[CheckoutOrderLogic] queryExchangeRate error", logger.Field("error", err.Error()))
		return "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "queryExchangeRate error: %s", err.Error())
	}
	convertAmount := int64(amount * 100)
	// create payment
	QRCode, err := client.PreCreateTrade(l.ctx, alipay.Order{
		OrderNo: info.OrderNo,
		Amount:  convertAmount,
	})
	if err != nil {
		l.Errorw("[CheckoutOrderLogic] PreCreateTrade error", logger.Field("error", err.Error()))
		return "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "PreCreateTrade error: %s", err.Error())
	}
	return QRCode, nil
}

// Stripe Payment
func (l *PurchaseCheckoutLogic) stripePayment(config string, info *order.Order, identifier string) (*types.StripePayment, error) {
	// stripe WeChat pay or stripe alipay
	stripeConfig := payment.StripeConfig{}
	if err := json.Unmarshal([]byte(config), &stripeConfig); err != nil {
		l.Errorw("[CheckoutOrderLogic] Unmarshal error", logger.Field("error", err.Error()))
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
		l.Errorw("[CheckoutOrderLogic] queryExchangeRate error", logger.Field("error", err.Error()))
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
			Email: identifier,
		})
	if err != nil {
		l.Errorw("[CheckoutOrderLogic] CreatePaymentSheet error", logger.Field("error", err.Error()))
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
		l.Errorw("[CheckoutOrderLogic] Update error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "Update error: %s", err.Error())
	}
	return stripePayment, nil
}

func (l *PurchaseCheckoutLogic) epayPayment(config *payment.Payment, info *order.Order, returnUrl string) (string, error) {
	epayConfig := payment.EPayConfig{}
	if err := json.Unmarshal([]byte(config.Config), &epayConfig); err != nil {
		l.Errorw("[CheckoutOrderLogic] Unmarshal error", logger.Field("error", err.Error()))
		return "", errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "Unmarshal error: %s", err.Error())
	}
	client := epay.NewClient(epayConfig.Pid, epayConfig.Url, epayConfig.Key)
	// Calculate the amount with exchange rate
	amount, err := l.queryExchangeRate("CNY", info.Amount)
	if err != nil {
		return "", err
	}
	notifyUrl := ""
	if config.Domain != "" {
		notifyUrl = config.Domain + "/v1/notify/" + config.Platform + "/" + config.Token
	} else {
		host, ok := l.ctx.Value(constant.CtxKeyRequestHost).(string)
		if !ok {
			host = l.svcCtx.Config.Host
		}
		notifyUrl = "https://" + host + "/v1/notify/" + config.Platform + "/" + config.Token
	}
	// create payment
	url := client.CreatePayUrl(epay.Order{
		Name:      l.svcCtx.Config.Site.SiteName,
		Amount:    amount,
		OrderNo:   info.OrderNo,
		SignType:  "MD5",
		NotifyUrl: notifyUrl,
		ReturnUrl: returnUrl,
	})
	return url, nil
}

// Query exchange rate
func (l *PurchaseCheckoutLogic) queryExchangeRate(to string, src int64) (amount float64, err error) {
	amount = float64(src) / float64(100)
	// query system currency
	currency, err := l.svcCtx.SystemModel.GetCurrencyConfig(l.ctx)
	if err != nil {
		l.Errorw("[CheckoutOrderLogic] GetCurrencyConfig error", logger.Field("error", err.Error()))
		return 0, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "GetCurrencyConfig error: %s", err.Error())
	}
	configs := struct {
		CurrencyUnit   string
		CurrencySymbol string
		AccessKey      string
	}{}
	tool.SystemConfigSliceReflectToStruct(currency, &configs)
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

// Balance payment
func (l *PurchaseCheckoutLogic) balancePayment(u *user.User, o *order.Order) error {
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
		l.Errorw("[CheckoutOrderLogic] Transaction error", logger.Field("error", err.Error()), logger.Field("orderNo", o.OrderNo))
		return err
	}
	// create activity order task
	payload := queueType.ForthwithActivateOrderPayload{
		OrderNo: o.OrderNo,
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		l.Errorw("[CheckoutOrderLogic] Marshal error", logger.Field("error", err.Error()))
		return err
	}

	task := asynq.NewTask(queueType.ForthwithActivateOrder, bytes)
	_, err = l.svcCtx.Queue.EnqueueContext(l.ctx, task)
	if err != nil {
		l.Errorw("[CheckoutOrderLogic] Enqueue error", logger.Field("error", err.Error()))
		return err
	}
	l.Logger.Info("[CheckoutOrderLogic] Enqueue success", logger.Field("orderNo", o.OrderNo))
	return nil
}
