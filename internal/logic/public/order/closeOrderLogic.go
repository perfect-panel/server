package order

import (
	"context"
	"encoding/json"
	"time"

	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/payment/stripe"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/payment"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/payment/alipay"
)

type CloseOrderLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCloseOrderLogic Close order
func NewCloseOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseOrderLogic {
	return &CloseOrderLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CloseOrderLogic) CloseOrder(req *types.CloseOrderRequest) error {
	// Find order information by order number
	orderInfo, err := l.svcCtx.OrderModel.FindOneByOrderNo(l.ctx, req.OrderNo)
	if err != nil {
		l.Errorw("[CloseOrder] Find order info failed",
			logger.Field("error", err.Error()),
			logger.Field("orderNo", req.OrderNo),
		)
		return nil
	}
	// If the order status is not 1, it means that the order has been closed or paid
	if orderInfo.Status != 1 {
		l.Infow("[CloseOrder] Order status is not 1",
			logger.Field("orderNo", req.OrderNo),
			logger.Field("status", orderInfo.Status),
		)
		return nil
	}
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// update order status
		err := tx.Model(&order.Order{}).Where("order_no = ?", req.OrderNo).Update("status", 3).Error
		if err != nil {
			l.Errorw("[CloseOrder] Update order status failed",
				logger.Field("error", err.Error()),
				logger.Field("orderNo", req.OrderNo),
			)
			return err
		}
		// If User ID is 0, it means that the order is a guest order and does not need to be refunded, the order can be deleted directly
		if orderInfo.UserId == 0 {
			err = tx.Model(&order.Order{}).Where("order_no = ?", req.OrderNo).Delete(&order.Order{}).Error
			if err != nil {
				l.Errorw("[CloseOrder] Delete order failed",
					logger.Field("error", err.Error()),
					logger.Field("orderNo", req.OrderNo),
				)
				return err
			}
			return nil
		}
		// refund deduction amount to user deduction balance
		if orderInfo.GiftAmount > 0 {
			userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, orderInfo.UserId)
			if err != nil {
				l.Errorw("[CloseOrder] Find user info failed",
					logger.Field("error", err.Error()),
					logger.Field("user_id", orderInfo.UserId),
				)
				return err
			}
			deduction := userInfo.GiftAmount + orderInfo.GiftAmount
			err = tx.Model(&user.User{}).Where("id = ?", orderInfo.UserId).Update("gift_amount", deduction).Error
			if err != nil {
				l.Errorw("[CloseOrder] Refund deduction amount failed",
					logger.Field("error", err.Error()),
					logger.Field("uid", orderInfo.UserId),
					logger.Field("deduction", orderInfo.GiftAmount),
				)
				return err
			}
			// Record the deduction refund log

			giftLog := log.Gift{
				Type:        log.GiftTypeIncrease,
				OrderNo:     orderInfo.OrderNo,
				SubscribeId: 0,
				Amount:      orderInfo.GiftAmount,
				Balance:     deduction,
				Remark:      "Order cancellation refund",
				Timestamp:   time.Now().UnixMilli(),
			}
			content, _ := giftLog.Marshal()

			err = tx.Model(&log.SystemLog{}).Create(&log.SystemLog{
				Id:       0,
				Type:     log.TypeGift.Uint8(),
				Date:     time.Now().Format(time.DateOnly),
				ObjectID: userInfo.Id,
				Content:  string(content),
			}).Error
			if err != nil {
				l.Errorw("[CloseOrder] Record cancellation refund log failed",
					logger.Field("error", err.Error()),
					logger.Field("uid", orderInfo.UserId),
					logger.Field("deduction", orderInfo.GiftAmount),
				)
				return err
			}
			// update user cache
			return l.svcCtx.UserModel.UpdateUserCache(l.ctx, userInfo)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// confirmationPayment Determine whether the payment is successful
//
//nolint:unused
func (l *CloseOrderLogic) confirmationPayment(order *order.Order) bool {
	paymentConfig, err := l.svcCtx.PaymentModel.FindOne(l.ctx, order.PaymentId)
	if err != nil {
		l.Errorw("[CloseOrder] Find payment config failed", logger.Field("error", err.Error()), logger.Field("paymentMark", order.Method))
		return false
	}
	switch order.Method {
	case AlipayF2f:
		if l.queryAlipay(paymentConfig, order.TradeNo) {
			return true
		}
	case StripeAlipay:
		if l.queryStripe(paymentConfig, order.TradeNo) {
			return true
		}
	case StripeWeChatPay:
		if l.queryStripe(paymentConfig, order.TradeNo) {
			return true
		}
	default:
		l.Infow("[CloseOrder] Unsupported payment method", logger.Field("paymentMethod", order.Method))
	}
	return false
}

// queryAlipay Query Alipay payment status
//
//nolint:unused
func (l *CloseOrderLogic) queryAlipay(paymentConfig *payment.Payment, TradeNo string) bool {
	config := payment.AlipayF2FConfig{}
	if err := json.Unmarshal([]byte(paymentConfig.Config), &config); err != nil {
		l.Errorw("[CloseOrder] Unmarshal payment config failed", logger.Field("error", err.Error()), logger.Field("config", paymentConfig.Config))
		return false
	}
	client := alipay.NewClient(alipay.Config{
		AppId:       config.AppId,
		PrivateKey:  config.PrivateKey,
		PublicKey:   config.PublicKey,
		InvoiceName: config.InvoiceName,
	})
	status, err := client.QueryTrade(l.ctx, TradeNo)
	if err != nil {
		l.Errorw("[CloseOrder] Query trade failed", logger.Field("error", err.Error()), logger.Field("TradeNo", TradeNo))
		return false
	}
	if status == alipay.Success || status == alipay.Finished {
		return true
	}
	return false
}

// queryStripe Query Stripe payment status
//
//nolint:unused
func (l *CloseOrderLogic) queryStripe(paymentConfig *payment.Payment, TradeNo string) bool {
	config := payment.StripeConfig{}
	if err := json.Unmarshal([]byte(paymentConfig.Config), &config); err != nil {
		l.Errorw("[CloseOrder] Unmarshal payment config failed", logger.Field("error", err.Error()), logger.Field("config", paymentConfig.Config))
		return false
	}
	client := stripe.NewClient(stripe.Config{
		PublicKey:     config.PublicKey,
		SecretKey:     config.SecretKey,
		WebhookSecret: config.WebhookSecret,
	})
	status, err := client.QueryOrderStatus(TradeNo)
	if err != nil {
		l.Errorw("[CloseOrder] Query order status failed", logger.Field("error", err.Error()), logger.Field("TradeNo", TradeNo))
		return false
	}
	return status
}
