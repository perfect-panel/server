// Package orderLogic provides order processing logic for handling various types of orders
// including subscription purchases, renewals, traffic resets, and balance recharges.
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
	"github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/uuidx"
	"github.com/perfect-panel/server/queue/types"
	"gorm.io/gorm"
)

// Order type constants define the different types of orders that can be processed
const (
	OrderTypeSubscribe    = 1 // New subscription purchase
	OrderTypeRenewal      = 2 // Subscription renewal
	OrderTypeResetTraffic = 3 // Traffic quota reset
	OrderTypeRecharge     = 4 // Balance recharge
)

// Order status constants define the lifecycle states of an order
const (
	OrderStatusPending  = 1 // Order created but not paid
	OrderStatusPaid     = 2 // Order paid and ready for processing
	OrderStatusClose    = 3 // Order closed/cancelled
	OrderStatusFailed   = 4 // Order processing failed
	OrderStatusFinished = 5 // Order successfully completed
)

// Commission type constants define the types of commission transactions
const (
	CommissionTypeRecharge = 1 // Commission from balance recharge
)

// Predefined error variables for common error conditions
var (
	ErrInvalidOrderStatus = fmt.Errorf("invalid order status")
	ErrInvalidOrderType   = fmt.Errorf("invalid order type")
)

// ActivateOrderLogic handles the activation and processing of paid orders
type ActivateOrderLogic struct {
	svc *svc.ServiceContext // Service context containing dependencies
}

// NewActivateOrderLogic creates a new instance of ActivateOrderLogic
func NewActivateOrderLogic(svc *svc.ServiceContext) *ActivateOrderLogic {
	return &ActivateOrderLogic{
		svc: svc,
	}
}

// ProcessTask is the main entry point for processing order activation tasks.
// It handles the complete workflow of activating a paid order including validation,
// processing based on order type, and finalization.
func (l *ActivateOrderLogic) ProcessTask(ctx context.Context, task *asynq.Task) error {
	payload, err := l.parsePayload(ctx, task.Payload())
	if err != nil {
		return nil // Log and continue
	}

	orderInfo, err := l.validateAndGetOrder(ctx, payload.OrderNo)
	if err != nil {
		return nil // Log and continue
	}

	if err := l.processOrderByType(ctx, orderInfo); err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Process task failed", logger.Field("error", err.Error()))
		return nil
	}

	l.finalizeCouponAndOrder(ctx, orderInfo)
	return nil
}

// parsePayload unmarshals the task payload into a structured format
func (l *ActivateOrderLogic) parsePayload(ctx context.Context, payload []byte) (*types.ForthwithActivateOrderPayload, error) {
	var p types.ForthwithActivateOrderPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.WithContext(ctx).Error("[ActivateOrderLogic] Unmarshal payload failed",
			logger.Field("error", err.Error()),
			logger.Field("payload", string(payload)),
		)
		return nil, err
	}
	return &p, nil
}

// validateAndGetOrder retrieves an order by order number and validates its status
// Returns error if order is not found or not in paid status
func (l *ActivateOrderLogic) validateAndGetOrder(ctx context.Context, orderNo string) (*order.Order, error) {
	orderInfo, err := l.svc.OrderModel.FindOneByOrderNo(ctx, orderNo)
	if err != nil {
		logger.WithContext(ctx).Error("Find order failed",
			logger.Field("error", err.Error()),
			logger.Field("order_no", orderNo),
		)
		return nil, err
	}

	if orderInfo.Status != OrderStatusPaid {
		logger.WithContext(ctx).Error("Order status error",
			logger.Field("order_no", orderInfo.OrderNo),
			logger.Field("status", orderInfo.Status),
		)
		return nil, ErrInvalidOrderStatus
	}

	return orderInfo, nil
}

// processOrderByType routes order processing based on the order type
func (l *ActivateOrderLogic) processOrderByType(ctx context.Context, orderInfo *order.Order) error {
	switch orderInfo.Type {
	case OrderTypeSubscribe:
		return l.NewPurchase(ctx, orderInfo)
	case OrderTypeRenewal:
		return l.Renewal(ctx, orderInfo)
	case OrderTypeResetTraffic:
		return l.ResetTraffic(ctx, orderInfo)
	case OrderTypeRecharge:
		return l.Recharge(ctx, orderInfo)
	default:
		logger.WithContext(ctx).Error("Order type is invalid", logger.Field("type", orderInfo.Type))
		return ErrInvalidOrderType
	}
}

// finalizeCouponAndOrder handles post-processing tasks including coupon updates
// and order status finalization
func (l *ActivateOrderLogic) finalizeCouponAndOrder(ctx context.Context, orderInfo *order.Order) {
	// Update coupon if exists
	if orderInfo.Coupon != "" {
		if err := l.svc.CouponModel.UpdateCount(ctx, orderInfo.Coupon); err != nil {
			logger.WithContext(ctx).Error("Update coupon status failed",
				logger.Field("error", err.Error()),
				logger.Field("coupon", orderInfo.Coupon),
			)
		}
	}

	// Update order status
	orderInfo.Status = OrderStatusFinished
	if err := l.svc.OrderModel.Update(ctx, orderInfo); err != nil {
		logger.WithContext(ctx).Error("Update order status failed",
			logger.Field("error", err.Error()),
			logger.Field("order_no", orderInfo.OrderNo),
		)
	}
}

// NewPurchase handles new subscription purchase including user creation,
// subscription setup, commission processing, cache updates, and notifications
func (l *ActivateOrderLogic) NewPurchase(ctx context.Context, orderInfo *order.Order) error {
	userInfo, err := l.getUserOrCreate(ctx, orderInfo)
	if err != nil {
		return err
	}

	sub, err := l.getSubscribeInfo(ctx, orderInfo.SubscribeId)
	if err != nil {
		return err
	}

	userSub, err := l.createUserSubscription(ctx, orderInfo, sub)
	if err != nil {
		return err
	}

	// Handle commission in separate goroutine to avoid blocking
	go l.handleCommission(context.Background(), userInfo, orderInfo, true)

	// Clear cache
	l.clearServerCache(ctx, sub)

	// Send notifications
	l.sendNotifications(ctx, orderInfo, userInfo, sub, userSub, telegram.PurchaseNotify)

	logger.WithContext(ctx).Info("Insert user subscribe success")
	return nil
}

// getUserOrCreate retrieves an existing user or creates a new guest user based on order details
func (l *ActivateOrderLogic) getUserOrCreate(ctx context.Context, orderInfo *order.Order) (*user.User, error) {
	if orderInfo.UserId != 0 {
		return l.getExistingUser(ctx, orderInfo.UserId)
	}
	return l.createGuestUser(ctx, orderInfo)
}

// getExistingUser retrieves user information by user ID
func (l *ActivateOrderLogic) getExistingUser(ctx context.Context, userId int64) (*user.User, error) {
	userInfo, err := l.svc.UserModel.FindOne(ctx, userId)
	if err != nil {
		logger.WithContext(ctx).Error("Find user failed",
			logger.Field("error", err.Error()),
			logger.Field("user_id", userId),
		)
		return nil, err
	}
	return userInfo, nil
}

// createGuestUser creates a new user account for guest orders using temporary order information
// stored in Redis cache
func (l *ActivateOrderLogic) createGuestUser(ctx context.Context, orderInfo *order.Order) (*user.User, error) {
	tempOrder, err := l.getTempOrderInfo(ctx, orderInfo.OrderNo)
	if err != nil {
		return nil, err
	}

	userInfo := &user.User{
		Password: tool.EncodePassWord(tempOrder.Password),
		AuthMethods: []user.AuthMethods{
			{
				AuthType:       tempOrder.AuthType,
				AuthIdentifier: tempOrder.Identifier,
			},
		},
	}

	err = l.svc.UserModel.Transaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Save(userInfo).Error; err != nil {
			return err
		}

		userInfo.ReferCode = uuidx.UserInviteCode(userInfo.Id)
		if err := tx.Model(&user.User{}).Where("id = ?", userInfo.Id).Update("refer_code", userInfo.ReferCode).Error; err != nil {
			return err
		}

		orderInfo.UserId = userInfo.Id
		return tx.Model(&order.Order{}).Where("order_no = ?", orderInfo.OrderNo).Update("user_id", userInfo.Id).Error
	})

	if err != nil {
		logger.WithContext(ctx).Error("Create user failed", logger.Field("error", err.Error()))
		return nil, err
	}

	// Handle referrer relationship
	l.handleReferrer(ctx, userInfo, tempOrder.InviteCode)

	logger.WithContext(ctx).Info("Create guest user success",
		logger.Field("user_id", userInfo.Id),
		logger.Field("identifier", tempOrder.Identifier),
		logger.Field("auth_type", tempOrder.AuthType),
	)

	return userInfo, nil
}

// getTempOrderInfo retrieves temporary order information from Redis cache
func (l *ActivateOrderLogic) getTempOrderInfo(ctx context.Context, orderNo string) (*constant.TemporaryOrderInfo, error) {
	cacheKey := fmt.Sprintf(constant.TempOrderCacheKey, orderNo)
	data, err := l.svc.Redis.Get(ctx, cacheKey).Result()
	if err != nil {
		logger.WithContext(ctx).Error("Get temp order cache failed",
			logger.Field("error", err.Error()),
			logger.Field("cache_key", cacheKey),
		)
		return nil, err
	}

	var tempOrder constant.TemporaryOrderInfo
	if err = json.Unmarshal([]byte(data), &tempOrder); err != nil {
		logger.WithContext(ctx).Error("Unmarshal temp order failed", logger.Field("error", err.Error()))
		return nil, err
	}

	return &tempOrder, nil
}

// handleReferrer establishes referrer relationship if an invite code is provided
func (l *ActivateOrderLogic) handleReferrer(ctx context.Context, userInfo *user.User, inviteCode string) {
	if inviteCode == "" {
		return
	}

	referer, err := l.svc.UserModel.FindOneByReferCode(ctx, inviteCode)
	if err != nil {
		logger.WithContext(ctx).Error("Find referer failed",
			logger.Field("error", err.Error()),
			logger.Field("refer_code", inviteCode),
		)
		return
	}

	userInfo.RefererId = referer.Id
	if err = l.svc.UserModel.Update(ctx, userInfo); err != nil {
		logger.WithContext(ctx).Error("Update user referer failed",
			logger.Field("error", err.Error()),
			logger.Field("user_id", userInfo.Id),
		)
	}
}

// getSubscribeInfo retrieves subscription plan details by subscription ID
func (l *ActivateOrderLogic) getSubscribeInfo(ctx context.Context, subscribeId int64) (*subscribe.Subscribe, error) {
	sub, err := l.svc.SubscribeModel.FindOne(ctx, subscribeId)
	if err != nil {
		logger.WithContext(ctx).Error("Find subscribe failed",
			logger.Field("error", err.Error()),
			logger.Field("subscribe_id", subscribeId),
		)
		return nil, err
	}
	return sub, nil
}

// createUserSubscription creates a new user subscription record based on order and subscription plan details
func (l *ActivateOrderLogic) createUserSubscription(ctx context.Context, orderInfo *order.Order, sub *subscribe.Subscribe) (*user.Subscribe, error) {
	now := time.Now()
	userSub := &user.Subscribe{
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

	if err := l.svc.UserModel.InsertSubscribe(ctx, userSub); err != nil {
		logger.WithContext(ctx).Error("Insert user subscribe failed", logger.Field("error", err.Error()))
		return nil, err
	}

	return userSub, nil
}

// handleCommission processes referral commission for the referrer if applicable.
// This runs asynchronously to avoid blocking the main order processing flow.
func (l *ActivateOrderLogic) handleCommission(ctx context.Context, userInfo *user.User, orderInfo *order.Order, isNewPurchase bool) {
	if !l.shouldProcessCommission(userInfo, orderInfo, isNewPurchase) {
		return
	}

	referer, err := l.svc.UserModel.FindOne(ctx, userInfo.RefererId)
	if err != nil {
		logger.WithContext(ctx).Error("Find referer failed",
			logger.Field("error", err.Error()),
			logger.Field("referer_id", userInfo.RefererId),
		)
		return
	}

	amount := l.calculateCommission(orderInfo.Price)

	// Use transaction for commission updates
	err = l.svc.DB.Transaction(func(tx *gorm.DB) error {
		referer.Commission += amount
		if err := l.svc.UserModel.Update(ctx, referer, tx); err != nil {
			return err
		}

		commissionLog := &user.CommissionLog{
			UserId:  referer.Id,
			OrderNo: orderInfo.OrderNo,
			Amount:  amount,
		}
		return l.svc.UserModel.InsertCommissionLog(ctx, commissionLog, tx)
	})

	if err != nil {
		logger.WithContext(ctx).Error("Update referer commission failed", logger.Field("error", err.Error()))
		return
	}

	// Update cache
	if err := l.svc.UserModel.UpdateUserCache(ctx, referer); err != nil {
		logger.WithContext(ctx).Error("Update referer cache failed",
			logger.Field("error", err.Error()),
			logger.Field("user_id", referer.Id),
		)
	}
}

// shouldProcessCommission determines if commission should be processed based on
// referrer existence, commission settings, and order type
func (l *ActivateOrderLogic) shouldProcessCommission(userInfo *user.User, orderInfo *order.Order, isNewPurchase bool) bool {
	return userInfo.RefererId != 0 &&
		l.svc.Config.Invite.ReferralPercentage != 0 &&
		(!l.svc.Config.Invite.OnlyFirstPurchase || (isNewPurchase && orderInfo.IsNew))
}

// calculateCommission computes the commission amount based on order price and referral percentage
func (l *ActivateOrderLogic) calculateCommission(price int64) int64 {
	return int64(float64(price) * (float64(l.svc.Config.Invite.ReferralPercentage) / 100))
}

// clearServerCache clears user list cache for all servers associated with the subscription
func (l *ActivateOrderLogic) clearServerCache(ctx context.Context, sub *subscribe.Subscribe) {
	serverIds := tool.StringToInt64Slice(sub.Server)
	groupServerIds := l.getServerIdsByGroups(ctx, sub.ServerGroup)
	allServerIds := append(serverIds, groupServerIds...)

	for _, id := range allServerIds {
		cacheKey := fmt.Sprintf("%s%d", config.ServerUserListCacheKey, id)
		if err := l.svc.Redis.Del(ctx, cacheKey).Err(); err != nil {
			logger.WithContext(ctx).Error("Del server user list cache failed",
				logger.Field("error", err.Error()),
				logger.Field("cache_key", cacheKey),
			)
		}
	}
}

// getServerIdsByGroups retrieves server IDs from server groups
func (l *ActivateOrderLogic) getServerIdsByGroups(ctx context.Context, serverGroup string) []int64 {
	data, err := l.svc.ServerModel.FindServerListByGroupIds(ctx, tool.StringToInt64Slice(serverGroup))
	if err != nil {
		logger.WithContext(ctx).Error("Find server list failed", logger.Field("error", err.Error()))
		return nil
	}

	serverIds := make([]int64, len(data))
	for i, item := range data {
		serverIds[i] = item.Id
	}
	return serverIds
}

// Renewal handles subscription renewal including subscription extension,
// traffic reset (if configured), commission processing, and notifications
func (l *ActivateOrderLogic) Renewal(ctx context.Context, orderInfo *order.Order) error {
	userInfo, err := l.getExistingUser(ctx, orderInfo.UserId)
	if err != nil {
		return err
	}

	userSub, err := l.getUserSubscription(ctx, orderInfo.SubscribeToken)
	if err != nil {
		return err
	}

	sub, err := l.getSubscribeInfo(ctx, orderInfo.SubscribeId)
	if err != nil {
		return err
	}

	if err := l.updateSubscriptionForRenewal(ctx, userSub, sub, orderInfo); err != nil {
		return err
	}

	// Handle commission
	go l.handleCommission(context.Background(), userInfo, orderInfo, false)

	// Send notifications
	l.sendNotifications(ctx, orderInfo, userInfo, sub, userSub, telegram.RenewalNotify)

	return nil
}

// getUserSubscription retrieves user subscription by token
func (l *ActivateOrderLogic) getUserSubscription(ctx context.Context, token string) (*user.Subscribe, error) {
	userSub, err := l.svc.UserModel.FindOneSubscribeByToken(ctx, token)
	if err != nil {
		logger.WithContext(ctx).Error("Find user subscribe failed", logger.Field("error", err.Error()))
		return nil, err
	}
	return userSub, nil
}

// updateSubscriptionForRenewal updates subscription details for renewal including
// expiration time extension and traffic reset if configured
func (l *ActivateOrderLogic) updateSubscriptionForRenewal(ctx context.Context, userSub *user.Subscribe, sub *subscribe.Subscribe, orderInfo *order.Order) error {
	now := time.Now()
	if userSub.ExpireTime.Before(now) {
		userSub.ExpireTime = now
	}

	// Reset traffic if enabled
	if sub.RenewalReset != nil && *sub.RenewalReset {
		userSub.Download = 0
		userSub.Upload = 0
	}

	if userSub.FinishedAt != nil {
		userSub.FinishedAt = nil
	}

	userSub.ExpireTime = tool.AddTime(sub.UnitTime, orderInfo.Quantity, userSub.ExpireTime)
	userSub.Status = 1

	if err := l.svc.UserModel.UpdateSubscribe(ctx, userSub); err != nil {
		logger.WithContext(ctx).Error("Update user subscribe failed", logger.Field("error", err.Error()))
		return err
	}

	return nil
}

// ResetTraffic handles traffic quota reset for existing subscriptions
func (l *ActivateOrderLogic) ResetTraffic(ctx context.Context, orderInfo *order.Order) error {
	userInfo, err := l.getExistingUser(ctx, orderInfo.UserId)
	if err != nil {
		return err
	}

	userSub, err := l.getUserSubscription(ctx, orderInfo.SubscribeToken)
	if err != nil {
		return err
	}

	// Reset traffic
	userSub.Download = 0
	userSub.Upload = 0
	userSub.Status = 1

	if err := l.svc.UserModel.UpdateSubscribe(ctx, userSub); err != nil {
		logger.WithContext(ctx).Error("Update user subscribe failed", logger.Field("error", err.Error()))
		return err
	}

	sub, err := l.getSubscribeInfo(ctx, userSub.SubscribeId)
	if err != nil {
		return err
	}

	// Send notifications
	l.sendNotifications(ctx, orderInfo, userInfo, sub, userSub, telegram.ResetTrafficNotify)

	return nil
}

// Recharge handles balance recharge orders including balance updates,
// transaction logging, and notifications
func (l *ActivateOrderLogic) Recharge(ctx context.Context, orderInfo *order.Order) error {
	userInfo, err := l.getExistingUser(ctx, orderInfo.UserId)
	if err != nil {
		return err
	}

	// Update balance in transaction
	err = l.svc.DB.Transaction(func(tx *gorm.DB) error {
		userInfo.Balance += orderInfo.Price
		if err := l.svc.UserModel.Update(ctx, userInfo, tx); err != nil {
			return err
		}

		balanceLog := &user.BalanceLog{
			UserId:  orderInfo.UserId,
			Amount:  orderInfo.Price,
			Type:    CommissionTypeRecharge,
			OrderId: orderInfo.Id,
			Balance: userInfo.Balance,
		}
		return l.svc.UserModel.InsertBalanceLog(ctx, balanceLog, tx)
	})

	if err != nil {
		logger.WithContext(ctx).Error("Database transaction failed", logger.Field("error", err.Error()))
		return err
	}

	// Send notifications
	l.sendRechargeNotifications(ctx, orderInfo, userInfo)

	return nil
}

// sendNotifications sends both user and admin notifications for order completion
func (l *ActivateOrderLogic) sendNotifications(ctx context.Context, orderInfo *order.Order, userInfo *user.User, sub *subscribe.Subscribe, userSub *user.Subscribe, notifyType string) {
	// Send user notification
	if telegramId, ok := findTelegram(userInfo); ok {
		templateData := l.buildUserNotificationData(orderInfo, sub, userSub)
		if text, err := tool.RenderTemplateToString(notifyType, templateData); err == nil {
			l.sendUserNotifyWithTelegram(telegramId, text)
		}
	}

	// Send admin notification
	adminData := l.buildAdminNotificationData(orderInfo, sub)
	if text, err := tool.RenderTemplateToString(telegram.AdminOrderNotify, adminData); err == nil {
		l.sendAdminNotifyWithTelegram(ctx, text)
	}
}

// sendRechargeNotifications sends specific notifications for balance recharge orders
func (l *ActivateOrderLogic) sendRechargeNotifications(ctx context.Context, orderInfo *order.Order, userInfo *user.User) {
	// Send user notification
	if telegramId, ok := findTelegram(userInfo); ok {
		templateData := map[string]string{
			"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
			"PaymentMethod": orderInfo.Method,
			"Time":          orderInfo.CreatedAt.Format("2006-01-02 15:04:05"),
			"Balance":       fmt.Sprintf("%.2f", float64(userInfo.Balance)/100),
		}
		if text, err := tool.RenderTemplateToString(telegram.RechargeNotify, templateData); err == nil {
			l.sendUserNotifyWithTelegram(telegramId, text)
		}
	}

	// Send admin notification
	adminData := map[string]string{
		"OrderNo":       orderInfo.OrderNo,
		"TradeNo":       orderInfo.TradeNo,
		"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
		"SubscribeName": "余额充值",
		"OrderStatus":   "已支付",
		"OrderTime":     orderInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		"PaymentMethod": orderInfo.Method,
	}
	if text, err := tool.RenderTemplateToString(telegram.AdminOrderNotify, adminData); err == nil {
		l.sendAdminNotifyWithTelegram(ctx, text)
	}
}

// buildUserNotificationData creates template data for user notifications
func (l *ActivateOrderLogic) buildUserNotificationData(orderInfo *order.Order, sub *subscribe.Subscribe, userSub *user.Subscribe) map[string]string {
	data := map[string]string{
		"OrderNo":       orderInfo.OrderNo,
		"SubscribeName": sub.Name,
		"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
	}

	if userSub != nil {
		data["ExpireTime"] = userSub.ExpireTime.Format("2006-01-02 15:04:05")
		data["ResetTime"] = time.Now().Format("2006-01-02 15:04:05")
	}

	return data
}

// buildAdminNotificationData creates template data for admin notifications
func (l *ActivateOrderLogic) buildAdminNotificationData(orderInfo *order.Order, sub *subscribe.Subscribe) map[string]string {
	subscribeName := sub.Name
	if orderInfo.Type == OrderTypeResetTraffic {
		subscribeName = "流量重置"
	}

	return map[string]string{
		"OrderNo":       orderInfo.OrderNo,
		"TradeNo":       orderInfo.TradeNo,
		"SubscribeName": subscribeName,
		"OrderAmount":   fmt.Sprintf("%.2f", float64(orderInfo.Price)/100),
		"OrderStatus":   "已支付",
		"OrderTime":     orderInfo.CreatedAt.Format("2006-01-02 15:04:05"),
		"PaymentMethod": orderInfo.Method,
	}
}

// sendUserNotifyWithTelegram sends a notification message to a user via Telegram
func (l *ActivateOrderLogic) sendUserNotifyWithTelegram(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "markdown"
	if _, err := l.svc.TelegramBot.Send(msg); err != nil {
		logger.Error("Send telegram user message failed", logger.Field("error", err.Error()))
	}
}

// sendAdminNotifyWithTelegram sends a notification message to all admin users via Telegram
func (l *ActivateOrderLogic) sendAdminNotifyWithTelegram(ctx context.Context, text string) {
	admins, err := l.svc.UserModel.QueryAdminUsers(ctx)
	if err != nil {
		logger.WithContext(ctx).Error("Query admin users failed", logger.Field("error", err.Error()))
		return
	}

	for _, admin := range admins {
		if telegramId, ok := findTelegram(admin); ok {
			msg := tgbotapi.NewMessage(telegramId, text)
			msg.ParseMode = "markdown"
			if _, err := l.svc.TelegramBot.Send(msg); err != nil {
				logger.WithContext(ctx).Error("Send telegram admin message failed", logger.Field("error", err.Error()))
			}
		}
	}
}

// findTelegram extracts Telegram chat ID from user authentication methods.
// Returns the chat ID and a boolean indicating if Telegram auth was found.
func findTelegram(u *user.User) (int64, bool) {
	for _, item := range u.AuthMethods {
		if item.AuthType == "telegram" {
			if telegramId, err := strconv.ParseInt(item.AuthIdentifier, 10, 64); err == nil {
				return telegramId, true
			}
		}
	}
	return 0, false
}
