package orderLogic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/perfect-panel/server/pkg/constant"

	"github.com/perfect-panel/server/pkg/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/logic/telegram"
	"github.com/perfect-panel/server/internal/model/order"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/uuidx"
	"github.com/perfect-panel/server/queue/types"
	"gorm.io/gorm"
)

const (
	Subscribe    = 1
	Renewal      = 2
	ResetTraffic = 3
	Recharge     = 4
)

type ActivateOrderLogic struct {
	svc *svc.ServiceContext
}

func NewActivateOrderLogic(svc *svc.ServiceContext) *ActivateOrderLogic {
	return &ActivateOrderLogic{
		svc: svc,
	}
}

func (l *ActivateOrderLogic) ProcessTask(ctx context.Context, task *asynq.Task) error {
	payload := types.ForthwithActivateOrderPayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Unmarshal payload failed",
			logger.Field("error", err.Error()),
			logger.Field("payload", string(task.Payload())),
		)
		return nil
	}
	// Find order by order no
	orderInfo, err := l.svc.OrderModel.FindOneByOrderNo(ctx, payload.OrderNo)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Find order failed",
			logger.Field("error", err.Error()),
			logger.Field("order_no", payload.OrderNo),
		)
		return nil
	}

	if orderInfo.Status != 2 {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Order status error",
			logger.Field("order_no", orderInfo.OrderNo),
			logger.Field("status", orderInfo.Status),
		)
		return nil
	}
	switch orderInfo.Type {
	case Subscribe:
		err = l.NewPurchase(ctx, orderInfo)
	case Renewal:
		err = l.Renewal(ctx, orderInfo)
	case ResetTraffic:
		err = l.ResetTraffic(ctx, orderInfo)
	case Recharge:
		err = l.Recharge(ctx, orderInfo)
	default:
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Order type is invalid", logger.Field("type", orderInfo.Type))
	}
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Process task failed", logger.Field("error", err.Error()))
		return nil
	}
	// if coupon is not empty
	if orderInfo.Coupon != "" {
		// update coupon status
		err = l.svc.CouponModel.UpdateCount(ctx, orderInfo.Coupon)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Update coupon status failed",
				logger.Field("error", err.Error()),
				logger.Field("coupon", orderInfo.Coupon),
			)
		}
	}
	// update order status
	orderInfo.Status = 5
	err = l.svc.OrderModel.Update(ctx, orderInfo)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Update order status failed",
			logger.Field("error", err.Error()),
			logger.Field("order_no", orderInfo.OrderNo),
		)
	}

	return nil
}

// NewPurchase New purchase
func (l *ActivateOrderLogic) NewPurchase(ctx context.Context, orderInfo *order.Order) error {
	var userInfo *user.User
	var err error
	if orderInfo.UserId != 0 {
		// find user by user id
		userInfo, err = l.svc.UserModel.FindOne(ctx, orderInfo.UserId)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Find user failed",
				logger.Field("error", err.Error()),
				logger.Field("user_id", orderInfo.UserId),
				logger.Field("user_id", orderInfo.UserId),
			)
			return err
		}
	} else {
		// If User ID is 0, it means that the order is a guest order, need to create a new user
		// query info with redis
		cacheKey := fmt.Sprintf(constant.TempOrderCacheKey, orderInfo.OrderNo)
		data, err := l.svc.Redis.Get(ctx, cacheKey).Result()
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Get temp order cache failed",
				logger.Field("error", err.Error()),
				logger.Field("cache_key", cacheKey),
			)
			return err
		}
		var tempOrder constant.TemporaryOrderInfo
		if err = json.Unmarshal([]byte(data), &tempOrder); err != nil {
			logger.WithContext(ctx).Errorw("[ActivateOrderLogic] Unmarshal temp order failed",
				logger.Field("error", err.Error()),
			)
			return err
		}
		// create user

		userInfo = &user.User{
			Password: tool.EncodePassWord(tempOrder.Password),
			AuthMethods: []user.AuthMethods{
				{
					AuthType:       tempOrder.AuthType,
					AuthIdentifier: tempOrder.Identifier,
				},
			},
		}
		err = l.svc.UserModel.Transaction(ctx, func(tx *gorm.DB) error {
			// Save user information
			if err := tx.Save(userInfo).Error; err != nil {
				return err
			}
			// Generate ReferCode
			userInfo.ReferCode = uuidx.UserInviteCode(userInfo.Id)
			// Update ReferCode
			if err := tx.Model(&user.User{}).Where("id = ?", userInfo.Id).Update("refer_code", userInfo.ReferCode).Error; err != nil {
				return err
			}
			orderInfo.UserId = userInfo.Id
			return tx.Model(&order.Order{}).Where("order_no = ?", orderInfo.OrderNo).Update("user_id", userInfo.Id).Error
		})
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Create user failed",
				logger.Field("error", err.Error()),
			)
			return err
		}

		if tempOrder.InviteCode != "" {
			// find referer by refer code
			referer, err := l.svc.UserModel.FindOneByReferCode(ctx, tempOrder.InviteCode)
			if err != nil {
				logger.WithContext(ctx).Error("[ActivateOrderLogic] Find referer failed",
					logger.Field("error", err.Error()),
					logger.Field("refer_code", tempOrder.InviteCode),
				)
			} else {
				userInfo.RefererId = referer.Id
				err = l.svc.UserModel.Update(ctx, userInfo)
				if err != nil {
					logger.WithContext(ctx).Error("[ActivateOrderLogic] Update user referer failed",
						logger.Field("error", err.Error()),
						logger.Field("user_id", userInfo.Id),
					)
				}
			}
		}

		logger.WithContext(ctx).Info("[ActivateOrderLogic] Create guest user success", logger.Field("user_id", userInfo.Id), logger.Field("Identifier", tempOrder.Identifier), logger.Field("AuthType", tempOrder.AuthType))
	}
	// find subscribe by id
	sub, err := l.svc.SubscribeModel.FindOne(ctx, orderInfo.SubscribeId)
	if err != nil {
		logger.WithContext(ctx).Errorw("[ActivateOrderLogic] Find subscribe failed",
			logger.Field("error", err.Error()),
			logger.Field("subscribe_id", orderInfo.SubscribeId),
		)
		return err
	}
	// create user subscribe
	now := time.Now()

	userSub := user.Subscribe{
		Id:          0,
		UserId:      orderInfo.UserId,
		OrderId:     orderInfo.Id,
		SubscribeId: orderInfo.SubscribeId,
		StartTime:   now,
		ExpireTime:  tool.AddTime(sub.UnitTime, orderInfo.Quantity, now),
		Traffic:     sub.Traffic,
		Download:    0,
		Upload:      0,
		Token:       uuidx.SubscribeToken(orderInfo.OrderNo),
		UUID:        uuid.New().String(),
		Status:      1,
	}
	err = l.svc.UserModel.InsertSubscribe(ctx, &userSub)

	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Insert user subscribe failed",
			logger.Field("error", err.Error()),
		)
		return err
	}
	// handler 	commission
	if userInfo.RefererId != 0 &&
		l.svc.Config.Invite.ReferralPercentage != 0 &&
		(!l.svc.Config.Invite.OnlyFirstPurchase || orderInfo.IsNew) {
		referer, err := l.svc.UserModel.FindOne(ctx, userInfo.RefererId)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Find referer failed",
				logger.Field("error", err.Error()),
				logger.Field("referer_id", userInfo.RefererId),
			)
			goto updateCache
		}
		// calculate commission
		amount := float64(orderInfo.Price) * (float64(l.svc.Config.Invite.ReferralPercentage) / 100)
		referer.Commission += int64(amount)
		err = l.svc.UserModel.Update(ctx, referer)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Update referer commission failed",
				logger.Field("error", err.Error()),
			)
			goto updateCache
		}
		// create commission log
		commissionLog := user.CommissionLog{
			UserId:  referer.Id,
			OrderNo: orderInfo.OrderNo,
			Amount:  int64(amount),
		}
		err = l.svc.UserModel.InsertCommissionLog(ctx, &commissionLog)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Insert commission log failed",
				logger.Field("error", err.Error()),
			)
		}
		err = l.svc.UserModel.UpdateUserCache(ctx, referer)
		if err != nil {
			logger.WithContext(ctx).Errorw("[ActivateOrderLogic] Update referer cache", logger.Field("error", err.Error()), logger.Field("user_id", referer.Id))
		}
	}
updateCache:
	for _, id := range tool.StringToInt64Slice(sub.Server) {
		cacheKey := fmt.Sprintf("%s%d", config.ServerUserListCacheKey, id)
		err = l.svc.Redis.Del(ctx, cacheKey).Err()
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Del server user list cache failed",
				logger.Field("error", err.Error()),
				logger.Field("cache_key", cacheKey),
			)
		}
	}
	data, err := l.svc.ServerModel.FindServerListByGroupIds(ctx, tool.StringToInt64Slice(sub.ServerGroup))
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Find server list failed", logger.Field("error", err.Error()))
		return nil
	}
	for _, item := range data {
		cacheKey := fmt.Sprintf("%s%d", config.ServerUserListCacheKey, item.Id)
		err = l.svc.Redis.Del(ctx, cacheKey).Err()
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Del server user list cache failed",
				logger.Field("error", err.Error()),
				logger.Field("cache_key", cacheKey),
			)
		}
	}
	userTelegramChatId, ok := findTelegram(userInfo)

	// sendMessage To Telegram
	if ok {
		text, err := tool.RenderTemplateToString(telegram.PurchaseNotify, map[string]string{
			"OrderNo":       orderInfo.OrderNo,
			"SubscribeName": sub.Name,
			"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
			"ExpireTime":    userSub.ExpireTime.Format("2006-01-02 15:04:05"),
		})
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Render template failed",
				logger.Field("error", err.Error()),
			)
		}
		l.sendUserNotifyWithTelegram(userTelegramChatId, text)
	}
	// send message to admin
	text, err := tool.RenderTemplateToString(telegram.AdminOrderNotify, map[string]string{
		"OrderNo":       orderInfo.OrderNo,
		"TradeNo":       orderInfo.TradeNo,
		"SubscribeName": sub.Name,
		//"UserEmail":     userInfo.Email,
		"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
		"OrderStatus":   "已支付",
		"OrderTime":     orderInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		"PaymentMethod": orderInfo.Method,
	})
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Render AdminOrderNotify template failed",
			logger.Field("error", err.Error()),
		)
	}
	l.sendAdminNotifyWithTelegram(ctx, text)
	logger.WithContext(ctx).Info("[ActivateOrderLogic] Insert user subscribe success")
	return nil
}

// Renewal Renewal
func (l *ActivateOrderLogic) Renewal(ctx context.Context, orderInfo *order.Order) error {
	// find user by user id
	userInfo, err := l.svc.UserModel.FindOne(ctx, orderInfo.UserId)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Find user failed",
			logger.Field("error", err.Error()),
			logger.Field("user_id", orderInfo.UserId),
		)
		return err
	}
	// find user subscribe by subscribe token
	userSub, err := l.svc.UserModel.FindOneSubscribeByOrderId(ctx, orderInfo.ParentId)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Find user subscribe failed",
			logger.Field("error", err.Error()),
			logger.Field("order_id", orderInfo.Id),
		)
		return err
	}
	// find subscribe by id
	sub, err := l.svc.SubscribeModel.FindOne(ctx, orderInfo.SubscribeId)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Find subscribe failed",
			logger.Field("error", err.Error()),
			logger.Field("subscribe_id", orderInfo.SubscribeId),
			logger.Field("order_id", orderInfo.Id),
		)
		return err
	}
	now := time.Now()
	if userSub.ExpireTime.Before(now) {
		userSub.ExpireTime = now
	}

	// Check whether traffic reset on renewal is enabled
	if *sub.RenewalReset {
		userSub.Download = 0
		userSub.Upload = 0
	}
	if userSub.FinishedAt != nil {
		userSub.FinishedAt = nil
	}

	userSub.ExpireTime = tool.AddTime(sub.UnitTime, orderInfo.Quantity, userSub.ExpireTime)
	userSub.Status = 1
	// update user subscribe
	err = l.svc.UserModel.UpdateSubscribe(ctx, userSub)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Update user subscribe failed",
			logger.Field("error", err.Error()),
		)
		return err
	}
	// handler 	commission
	if userInfo.RefererId != 0 &&
		l.svc.Config.Invite.ReferralPercentage != 0 &&
		!l.svc.Config.Invite.OnlyFirstPurchase {
		referer, err := l.svc.UserModel.FindOne(ctx, userInfo.RefererId)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Find referer failed",
				logger.Field("error", err.Error()),
				logger.Field("referer_id", userInfo.RefererId),
			)
			goto sendMessage
		}
		// calculate commission
		amount := float64(orderInfo.Price) * (float64(l.svc.Config.Invite.ReferralPercentage) / 100)
		referer.Commission += int64(amount)
		err = l.svc.UserModel.Update(ctx, referer)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Update referer commission failed",
				logger.Field("error", err.Error()),
			)
			goto sendMessage
		}
		// create commission log
		commissionLog := user.CommissionLog{
			UserId:  referer.Id,
			OrderNo: orderInfo.OrderNo,
			Amount:  int64(amount),
		}
		err = l.svc.UserModel.InsertCommissionLog(ctx, &commissionLog)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Insert commission log failed",
				logger.Field("error", err.Error()),
			)
		}
		err = l.svc.UserModel.UpdateUserCache(ctx, referer)
		if err != nil {
			logger.WithContext(ctx).Errorw("[ActivateOrderLogic] Update referer cache", logger.Field("error", err.Error()), logger.Field("user_id", referer.Id))
		}
	}

sendMessage:
	userTelegramChatId, ok := findTelegram(userInfo)
	// SendMessage To Telegram
	if ok {
		text, err := tool.RenderTemplateToString(telegram.RenewalNotify, map[string]string{
			"OrderNo":       orderInfo.OrderNo,
			"SubscribeName": sub.Name,
			"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
			"ExpireTime":    userSub.ExpireTime.Format("2006-01-02 15:04:05"),
		})
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Render template failed",
				logger.Field("error", err.Error()),
			)
		}
		l.sendUserNotifyWithTelegram(userTelegramChatId, text)
	}

	// send message to admin
	text, err := tool.RenderTemplateToString(telegram.AdminOrderNotify, map[string]string{
		"OrderNo":       orderInfo.OrderNo,
		"TradeNo":       orderInfo.TradeNo,
		"SubscribeName": sub.Name,
		//"UserEmail":     userInfo.Email,
		"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
		"OrderStatus":   "已支付",
		"OrderTime":     orderInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		"PaymentMethod": orderInfo.Method,
	})
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Render AdminOrderNotify template failed",
			logger.Field("error", err.Error()),
		)
	}
	l.sendAdminNotifyWithTelegram(ctx, text)
	return nil
}

// ResetTraffic Reset traffic
func (l *ActivateOrderLogic) ResetTraffic(ctx context.Context, orderInfo *order.Order) error {
	// find user by user id
	userInfo, err := l.svc.UserModel.FindOne(ctx, orderInfo.UserId)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Find user failed",
			logger.Field("error", err.Error()),
			logger.Field("user_id", orderInfo.UserId),
		)
		return err
	}
	// Generate a Subscribe Token through orderNo
	// find user subscribe by subscribe token
	userSub, err := l.svc.UserModel.FindOneSubscribeByToken(ctx, orderInfo.SubscribeToken)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Find user subscribe failed",
			logger.Field("error", err.Error()),
			logger.Field("order_id", orderInfo.Id),
		)
		return err
	}
	userSub.Download = 0
	userSub.Upload = 0
	userSub.Status = 1
	// update user subscribe
	err = l.svc.UserModel.UpdateSubscribe(ctx, userSub)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Update user subscribe failed",
			logger.Field("error", err.Error()),
		)
		return err
	}
	sub, err := l.svc.SubscribeModel.FindOne(ctx, userSub.SubscribeId)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Find subscribe failed",
			logger.Field("error", err.Error()),
			logger.Field("subscribe_id", userSub.SubscribeId),
		)
		return nil
	}
	userTelegramChatId, ok := findTelegram(userInfo)
	// SendMessage To Telegram
	if ok {
		text, err := tool.RenderTemplateToString(telegram.ResetTrafficNotify, map[string]string{
			"OrderNo":       orderInfo.OrderNo,
			"SubscribeName": sub.Name,
			"ResetTime":     time.Now().Format("2006-01-02 15:04:05"),
			"ExpireTime":    userSub.ExpireTime.Format("2006-01-02 15:04:05"),
		})
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Render template failed",
				logger.Field("error", err.Error()),
			)
		}
		l.sendUserNotifyWithTelegram(userTelegramChatId, text)
	}

	// send message to admin
	text, err := tool.RenderTemplateToString(telegram.AdminOrderNotify, map[string]string{
		"OrderNo":       orderInfo.OrderNo,
		"TradeNo":       orderInfo.TradeNo,
		"SubscribeName": "流量重置",
		"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
		"OrderStatus":   "已支付",
		"OrderTime":     orderInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		"PaymentMethod": orderInfo.Method,
	})
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Render AdminOrderNotify template failed",
			logger.Field("error", err.Error()),
		)
	}
	l.sendAdminNotifyWithTelegram(ctx, text)
	return nil
}

// Recharge Recharge to user
func (l *ActivateOrderLogic) Recharge(ctx context.Context, orderInfo *order.Order) error {
	// find user by user id
	userInfo, err := l.svc.UserModel.FindOne(ctx, orderInfo.UserId)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Find user failed",
			logger.Field("error", err.Error()),
			logger.Field("user_id", orderInfo.UserId),
		)
		return err
	}
	userInfo.Balance += orderInfo.Price
	// update user
	err = l.svc.DB.Transaction(func(tx *gorm.DB) error {
		err = l.svc.UserModel.Update(ctx, userInfo, tx)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Update user failed",
				logger.Field("error", err.Error()),
			)
			return err
		}
		// Create Balance Log
		balanceLog := user.BalanceLog{
			UserId:  orderInfo.UserId,
			Amount:  orderInfo.Price,
			Type:    1,
			OrderId: orderInfo.Id,
			Balance: userInfo.Balance,
		}
		err = l.svc.UserModel.InsertBalanceLog(ctx, &balanceLog, tx)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Insert balance log failed",
				logger.Field("error", err.Error()),
			)
			return err
		}

		return nil
	})
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Database transaction failed",
			logger.Field("error", err.Error()),
		)
		return err
	}
	userTelegramChatId, ok := findTelegram(userInfo)
	// SendMessage To Telegram
	if ok {
		text, err := tool.RenderTemplateToString(telegram.RechargeNotify, map[string]string{
			"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
			"PaymentMethod": orderInfo.Method,
			"Time":          orderInfo.CreatedAt.Format("2006-01-02 15:04:05"),
			"Balance":       fmt.Sprintf("%.2f", float64(userInfo.Balance)/100),
		})
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Render template failed",
				logger.Field("error", err.Error()),
			)
		}
		l.sendUserNotifyWithTelegram(userTelegramChatId, text)
	}
	// send message to admin
	text, err := tool.RenderTemplateToString(telegram.AdminOrderNotify, map[string]string{
		"OrderNo":       orderInfo.OrderNo,
		"TradeNo":       orderInfo.TradeNo,
		"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
		"SubscribeName": "余额充值",
		"OrderStatus":   "已支付",
		"OrderTime":     orderInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		"PaymentMethod": orderInfo.Method,
	})
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Render AdminOrderNotify template failed",
			logger.Field("error", err.Error()),
		)
	}
	l.sendAdminNotifyWithTelegram(ctx, text)
	return nil
}

// sendUserNotifyWithTelegram send message to user
func (l *ActivateOrderLogic) sendUserNotifyWithTelegram(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "markdown"
	_, err := l.svc.TelegramBot.Send(msg)
	if err != nil {
		logger.Error("[ActivateOrderLogic] Send telegram user message failed",
			logger.Field("error", err.Error()),
		)
	}
}

// sendAdminNotifyWithTelegram send message to admin
func (l *ActivateOrderLogic) sendAdminNotifyWithTelegram(ctx context.Context, text string) {
	admins, err := l.svc.UserModel.QueryAdminUsers(ctx)
	if err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Query admin users failed",
			logger.Field("error", err.Error()),
		)
		return
	}
	for _, admin := range admins {
		telegramId, ok := findTelegram(admin)
		if !ok {
			continue
		}
		msg := tgbotapi.NewMessage(telegramId, text)
		msg.ParseMode = "markdown"
		_, err := l.svc.TelegramBot.Send(msg)
		if err != nil {
			logger.WithContext(ctx).Error("[ActivateOrderLogic] Send telegram admin message failed",
				logger.Field("error", err.Error()),
			)
		}
	}
}

// findTelegram find user telegram id
func findTelegram(u *user.User) (int64, bool) {
	for _, item := range u.AuthMethods {
		if item.AuthType == "telegram" {
			// string to int64
			parseInt, err := strconv.ParseInt(item.AuthIdentifier, 10, 64)
			if err != nil {
				return 0, false
			}
			return parseInt, true
		}

	}
	return 0, false
}
