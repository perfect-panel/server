package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/log"
	"github.com/perfect-panel/server/internal/model/task"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"gorm.io/gorm"
)

const (
	UnitTimeNoLimit = "NoLimit" // Unlimited time subscription
	UnitTimeYear    = "Year"    // Annual subscription
	UnitTimeMonth   = "Month"   // Monthly subscription
	UnitTimeDay     = "Day"     // Daily subscription
	UnitTimeHour    = "Hour"    // Hourly subscription
	UnitTimeMinute  = "Minute"  // Per-minute subscription

)

type QuotaTaskLogic struct {
	svcCtx *svc.ServiceContext
}

type ErrorInfo struct {
	UserSubscribeId int64  `json:"user_subscribe_id"`
	Error           string `json:"error"`
}

func NewQuotaTaskLogic(svcCtx *svc.ServiceContext) *QuotaTaskLogic {
	return &QuotaTaskLogic{
		svcCtx: svcCtx,
	}
}

func (l *QuotaTaskLogic) ProcessTask(ctx context.Context, t *asynq.Task) error {
	taskID, err := l.parseTaskID(ctx, t.Payload())
	if err != nil {
		return err
	}

	taskInfo, err := l.getTaskInfo(ctx, taskID)
	if err != nil {
		return err
	}

	if taskInfo.Status != 0 {
		logger.WithContext(ctx).Info("[QuotaTaskLogic.ProcessTask] task already processed",
			logger.Field("taskID", taskID),
			logger.Field("status", taskInfo.Status),
		)
		return nil
	}

	scope, content, err := l.parseTaskData(ctx, taskInfo)
	if err != nil {
		return err
	}

	subscribes, err := l.getSubscribes(ctx, scope.Objects)
	if err != nil {
		return err
	}
	if err = l.processSubscribes(ctx, subscribes, content, taskInfo); err != nil {
		return err
	}
	// 清理用户缓存（仅在有赠送金时清理）
	if content.GiftValue != 0 {
		var userIds []int64
		for _, sub := range subscribes {
			userIds = append(userIds, sub.UserId)
		}
		userIds = tool.RemoveDuplicateElements(userIds...)
		var users []*user.User
		if err = l.svcCtx.DB.WithContext(ctx).Model(&user.User{}).Where("id IN ?", userIds).Find(&users).Error; err != nil {
			logger.WithContext(ctx).Error("[QuotaTaskLogic.ProcessTask] find users error",
				logger.Field("error", err.Error()),
				logger.Field("userIDs", userIds))
		}
		err = l.svcCtx.UserModel.ClearUserCache(ctx, users...)
		if err != nil {
			logger.WithContext(ctx).Error("[QuotaTaskLogic.ProcessTask] clear user cache error",
				logger.Field("error", err.Error()),
				logger.Field("userIDs", userIds))
		}
	}

	// 清理用户订阅缓存
	err = l.svcCtx.UserModel.ClearSubscribeCache(ctx, subscribes...)
	if err != nil {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.ProcessTask] clear subscribe cache error",
			logger.Field("error", err.Error()))
	}

	return nil
}

func (l *QuotaTaskLogic) parseTaskID(ctx context.Context, payload []byte) (int64, error) {
	if len(payload) == 0 {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.parseTaskID] empty payload")
		return 0, asynq.SkipRetry
	}

	taskID, err := strconv.ParseInt(string(payload), 10, 64)
	if err != nil {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.parseTaskID] invalid task ID",
			logger.Field("error", err.Error()),
			logger.Field("payload", string(payload)),
		)
		return 0, asynq.SkipRetry
	}
	return taskID, nil
}

func (l *QuotaTaskLogic) getTaskInfo(ctx context.Context, taskID int64) (*task.Task, error) {
	var taskInfo *task.Task
	if err := l.svcCtx.DB.WithContext(ctx).Model(&task.Task{}).Where("id = ?", taskID).First(&taskInfo).Error; err != nil {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.getTaskInfo] find task error",
			logger.Field("error", err.Error()),
			logger.Field("taskID", taskID),
		)
		return nil, asynq.SkipRetry
	}
	return taskInfo, nil
}

func (l *QuotaTaskLogic) parseTaskData(ctx context.Context, taskInfo *task.Task) (task.QuotaScope, task.QuotaContent, error) {
	var scope task.QuotaScope
	if err := scope.Unmarshal([]byte(taskInfo.Scope)); err != nil {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.parseTaskData] unmarshal scope error",
			logger.Field("error", err.Error()),
		)
		return scope, task.QuotaContent{}, asynq.SkipRetry
	}

	var content task.QuotaContent
	if err := content.Unmarshal([]byte(taskInfo.Content)); err != nil {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.parseTaskData] unmarshal content error",
			logger.Field("error", err.Error()),
		)
		return scope, content, asynq.SkipRetry
	}
	return scope, content, nil
}

func (l *QuotaTaskLogic) getSubscribes(ctx context.Context, subscriberIDs []int64) ([]*user.Subscribe, error) {
	var subscribes []*user.Subscribe
	if err := l.svcCtx.DB.WithContext(ctx).Model(&user.Subscribe{}).Where("id IN ?", subscriberIDs).Find(&subscribes).Error; err != nil {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.getSubscribes] find subscribes error",
			logger.Field("error", err.Error()),
			logger.Field("subscribers", subscriberIDs),
		)
		return nil, asynq.SkipRetry
	}
	return subscribes, nil
}

func (l *QuotaTaskLogic) processSubscribes(ctx context.Context, subscribes []*user.Subscribe, content task.QuotaContent, taskInfo *task.Task) error {
	tx := l.svcCtx.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.WithContext(ctx).Error("[QuotaTaskLogic.processSubscribes] transaction panic",
				logger.Field("panic", r),
			)
		}
	}()

	var errors []ErrorInfo
	now := time.Now()

	for _, sub := range subscribes {
		if err := l.processSubscription(tx, sub, content, now, &errors); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 根据错误情况决定任务状态
	status := int8(2) // Completed
	if len(errors) > 0 {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.processSubscribes] some subscriptions failed",
			logger.Field("total", len(subscribes)),
			logger.Field("failed", len(errors)),
		)
		// 如果所有订阅都失败，标记为失败状态
		if len(errors) == len(subscribes) {
			status = 3 // Failed
		}
		errs, err := json.Marshal(errors)
		if err != nil {
			logger.WithContext(ctx).Error("[QuotaTaskLogic.processSubscribes] marshal errors failed",
				logger.Field("error", err.Error()),
			)
			tx.Rollback()
			return err
		}
		taskInfo.Errors = string(errs)
	}

	taskInfo.Current = uint64(len(subscribes))
	taskInfo.Status = status
	err := tx.Where("id = ?", taskInfo.Id).Save(taskInfo).Error
	if err != nil {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.processSubscribes] update task status error",
			logger.Field("error", err.Error()),
			logger.Field("taskID", taskInfo.Id),
		)
		tx.Rollback()
		return err
	}

	if err = tx.Commit().Error; err != nil {
		logger.WithContext(ctx).Error("[QuotaTaskLogic.processSubscribes] commit transaction error",
			logger.Field("error", err.Error()),
		)
		return err
	}

	return nil
}

func (l *QuotaTaskLogic) processSubscription(tx *gorm.DB, sub *user.Subscribe, content task.QuotaContent, now time.Time, errors *[]ErrorInfo) error {
	// 验证订阅数据
	if sub == nil {
		*errors = append(*errors, ErrorInfo{
			UserSubscribeId: 0,
			Error:           "subscription is nil",
		})
		return nil
	}

	updated := false

	// 处理时间延长 - 修复逻辑：只要Days不为0就处理，不管ExpireTime是否为0
	if content.Days != 0 {
		if sub.ExpireTime.Unix() == 0 || sub.ExpireTime.Before(now) {
			// 如果没有过期时间或已过期，从现在开始计算
			sub.ExpireTime = now.AddDate(0, 0, int(content.Days))
		} else {
			// 在原有过期时间基础上延长
			sub.ExpireTime = sub.ExpireTime.AddDate(0, 0, int(content.Days))
		}
		// 如果订阅延长到未来时间，设置为激活状态
		if sub.ExpireTime.After(now) && sub.Status != 1 {
			sub.Status = 1 // Active
		}
		updated = true
	}

	// 处理流量重置
	if content.ResetTraffic {
		sub.Download = 0
		sub.Upload = 0
		updated = true
		if err := l.createResetTrafficLog(tx, sub.Id, sub.UserId, now); err != nil {
			// 记录错误但不阻断整个任务,日志失败不影响主流程
			*errors = append(*errors, ErrorInfo{
				UserSubscribeId: sub.Id,
				Error:           "create reset traffic log error: " + err.Error(),
			})
		}
	}

	// 处理赠送金
	if content.GiftValue != 0 {
		if err := l.processGift(tx, sub, content, now, errors); err != nil {
			return err
		}
	}

	// 只有在有更新时才保存订阅信息
	if updated {
		if err := tx.Where("id = ?", sub.Id).Save(sub).Error; err != nil {
			*errors = append(*errors, ErrorInfo{
				UserSubscribeId: sub.Id,
				Error:           "update subscription error: " + err.Error(),
			})
			return nil
		}
	}

	return nil
}

func (l *QuotaTaskLogic) processGift(tx *gorm.DB, sub *user.Subscribe, content task.QuotaContent, now time.Time, errors *[]ErrorInfo) error {
	// 验证赠送类型
	if content.GiftType != 1 && content.GiftType != 2 {
		*errors = append(*errors, ErrorInfo{
			UserSubscribeId: sub.Id,
			Error:           fmt.Sprintf("invalid gift type: %d", content.GiftType),
		})
		return nil
	}

	var userInfo user.User
	if err := tx.Model(&user.User{}).Where("id = ?", sub.UserId).First(&userInfo).Error; err != nil {
		*errors = append(*errors, ErrorInfo{
			UserSubscribeId: sub.Id,
			Error:           "find user error: " + err.Error(),
		})
		return nil
	}

	var giftAmount int64
	switch content.GiftType {
	case 1:
		giftAmount = int64(content.GiftValue)
	case 2:
		// 获取订阅对应的套餐信息
		subscribeInfo, err := l.svcCtx.SubscribeModel.FindOne(context.Background(), sub.SubscribeId)
		if err != nil {
			*errors = append(*errors, ErrorInfo{
				UserSubscribeId: sub.Id,
				Error:           "find subscribe error: " + err.Error(),
			})
			return nil
		}
		if subscribeInfo.UnitPrice > 0 {
			giftAmount = int64(float64(subscribeInfo.UnitPrice) * (float64(content.GiftValue) / 100))
		}
	}

	if giftAmount > 0 {
		userInfo.GiftAmount += giftAmount
		// 使用Update而不是Save，更精确地更新单个字段
		if err := tx.Model(&user.User{}).Where("id = ?", sub.UserId).Update("gift_amount", userInfo.GiftAmount).Error; err != nil {
			*errors = append(*errors, ErrorInfo{
				UserSubscribeId: sub.Id,
				Error:           "update user gift amount error: " + err.Error(),
			})
			return nil
		}

		if err := l.createGiftLog(tx, sub.Id, userInfo.Id, giftAmount, userInfo.GiftAmount, now); err != nil {
			*errors = append(*errors, ErrorInfo{
				UserSubscribeId: sub.Id,
				Error:           "create gift log error: " + err.Error(),
			})
			// 回滚用户金额更新
			userInfo.GiftAmount -= giftAmount
			tx.Model(&user.User{}).Where("id = ?", sub.UserId).Update("gift_amount", userInfo.GiftAmount)
			return nil
		}
	}

	return nil
}

func (l *QuotaTaskLogic) getStartTime(sub *user.Subscribe, now time.Time) time.Time {
	if sub.StartTime.Unix() == 0 {
		return now
	}
	return sub.StartTime
}

func (l *QuotaTaskLogic) createGiftLog(tx *gorm.DB, subscribeId, userId, amount, balance int64, now time.Time) error {
	giftLog := &log.Gift{
		Type:        log.GiftTypeIncrease,
		OrderNo:     "",
		SubscribeId: subscribeId,
		Amount:      amount,
		Balance:     balance,
		Remark:      "Quota task gift",
		Timestamp:   now.UnixMilli(),
	}

	logString, err := giftLog.Marshal()
	if err != nil {
		return fmt.Errorf("marshal gift log error: %v", err)
	}
	return tx.Model(&log.SystemLog{}).Create(&log.SystemLog{
		Type:     log.TypeGift.Uint8(),
		Content:  string(logString),
		ObjectID: userId,
		Date:     now.Format(time.DateOnly),
	}).Error
}

func (l *QuotaTaskLogic) createResetTrafficLog(tx *gorm.DB, subscribeId, userId int64, now time.Time) error {
	trafficLog := &log.ResetSubscribe{
		Type:      log.ResetSubscribeTypeQuota,
		UserId:    userId,
		OrderNo:   "",
		Timestamp: now.UnixMilli(),
	}

	logString, err := trafficLog.Marshal()
	if err != nil {
		return fmt.Errorf("marshal traffic log error: %v", err)
	}
	return tx.Model(&log.SystemLog{}).Create(&log.SystemLog{
		Type:     log.TypeResetSubscribe.Uint8(),
		Content:  string(logString),
		ObjectID: subscribeId,
		Date:     now.Format(time.DateOnly),
	}).Error
}
