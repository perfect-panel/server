package subscription

import (
	"context"
	"encoding/json"
	"time"

	queue "github.com/perfect-panel/server/queue/types"

	"github.com/perfect-panel/server/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"gorm.io/gorm"
)

type CheckSubscriptionLogic struct {
	svc *svc.ServiceContext
}

func NewCheckSubscriptionLogic(svc *svc.ServiceContext) *CheckSubscriptionLogic {
	return &CheckSubscriptionLogic{
		svc: svc,
	}
}

func (l *CheckSubscriptionLogic) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	logger.Infof("[CheckSubscription] Start check subscription: %s", time.Now().Format("2006-01-02 15:04:05"))
	// Check subscription traffic
	err := l.svc.UserModel.Transaction(ctx, func(db *gorm.DB) error {
		var list []*user.Subscribe
		err := db.Model(&user.Subscribe{}).Where("upload + download >= traffic AND status IN (0, 1)  AND traffic > 0 ").Find(&list).Error
		if err != nil {
			logger.Errorw("[Check Subscription Traffic] Query subscribe failed", logger.Field("error", err.Error()))
			return err
		}
		var ids []int64
		for _, item := range list {
			ids = append(ids, item.Id)
		}
		if len(ids) > 0 {
			err = db.Model(&user.Subscribe{}).Where("id IN ?", ids).Updates(map[string]interface{}{
				"status":      2,
				"finished_at": time.Now(),
			}).Error
			if err != nil {
				logger.Errorw("[Check Subscription Traffic] Update subscribe status failed", logger.Field("error", err.Error()))
				return nil
			}
			err = l.sendTrafficNotify(ctx, ids)
			if err != nil {
				logger.Errorw("[Check Subscription Traffic] Send email failed", logger.Field("error", err.Error()))
				return nil
			}

			if len(list) > 0 {
				if err = l.svc.UserModel.ClearSubscribeCache(ctx, list...); err != nil {
					logger.Errorw("[Check Subscription Traffic] Clear subscribe cache failed", logger.Field("error", err.Error()))
					return err
				}
			}
			l.clearServerCache(ctx, list...)
			logger.Infow("[Check Subscription Traffic] Update subscribe status", logger.Field("user_ids", ids), logger.Field("count", int64(len(ids))))

		} else {
			logger.Info("[Check Subscription Traffic] No subscribe need to update")
		}

		return nil
	})
	if err != nil {
		logger.Error("[CheckSubscription] Transaction failed", logger.Field("error", err.Error()))
	}
	// Check subscription expire
	err = l.svc.UserModel.Transaction(ctx, func(db *gorm.DB) error {
		var list []*user.Subscribe
		err = db.Model(&user.Subscribe{}).Where("`status` IN (0, 1) AND `expire_time` < ? AND `expire_time` != ? and `finished_at` IS NULL", time.Now(), time.UnixMilli(0)).Find(&list).Error
		if err != nil {
			logger.Error("[Check Subscription] Find subscribe failed", logger.Field("error", err.Error()))
			return err
		}
		var ids []int64
		for _, item := range list {
			ids = append(ids, item.Id)
		}
		if len(ids) > 0 {
			err = db.Model(&user.Subscribe{}).Where("id IN ?", ids).Updates(map[string]interface{}{
				"status":      3,
				"finished_at": time.Now(),
			}).Error
			if err != nil {
				logger.Error("[Check Subscription Expire] Update subscribe status failed", logger.Field("error", err.Error()))
				return err
			}
			err = l.sendExpiredNotify(ctx, ids)
			if err != nil {
				logger.Error("[Check Subscription Expire] Send email failed", logger.Field("error", err.Error()))
				return nil
			}
			if err = l.svc.UserModel.ClearSubscribeCache(ctx, list...); err != nil {
				logger.Errorw("[Check Subscription Traffic] Clear subscribe cache failed", logger.Field("error", err.Error()))
				return err
			}
			l.clearServerCache(ctx, list...)

			logger.Info("[Check Subscription Expire] Update subscribe status", logger.Field("user_ids", ids), logger.Field("count", int64(len(ids))))
		} else {
			logger.Info("[Check Subscription Expire] No subscribe need to update")
		}
		return nil
	})
	if err != nil {
		logger.Info("[CheckSubscription] Transaction failed", logger.Field("error", err.Error()))
	}
	return nil
}

func (l *CheckSubscriptionLogic) sendExpiredNotify(ctx context.Context, subs []int64) error {
	for _, id := range subs {
		sub, err := l.svc.UserModel.FindOneUserSubscribe(ctx, id)
		if err != nil {
			logger.Errorw("[CheckSubscription] FindOneUserSubscribe failed", logger.Field("error", err.Error()))
			continue
		}
		method, err := l.svc.UserModel.FindUserAuthMethodByUserId(ctx, "email", sub.UserId)
		if err != nil {
			logger.Errorw("[CheckSubscription] FindUserAuthMethodByUserId failed", logger.Field("error", err.Error()), logger.Field("user_id", sub.UserId))
			continue
		}
		var taskPayload queue.SendEmailPayload
		taskPayload.Type = queue.EmailTypeExpiration
		taskPayload.Email = method.AuthIdentifier
		taskPayload.Subject = "Subscription Expired"
		taskPayload.Content = map[string]interface{}{
			"SiteLogo":   l.svc.Config.Site.SiteLogo,
			"SiteName":   l.svc.Config.Site.SiteName,
			"ExpireDate": sub.ExpireTime.Format("2006-01-02 15:04:05"),
		}
		payloadBuy, err := json.Marshal(taskPayload)
		if err != nil {
			logger.Errorw("[CheckSubscription] Marshal payload failed", logger.Field("error", err.Error()))
			continue
		}
		task := asynq.NewTask(queue.ForthwithSendEmail, payloadBuy, asynq.MaxRetry(3))
		taskInfo, err := l.svc.Queue.Enqueue(task)
		if err != nil {
			logger.Errorw("[CheckSubscription] Enqueue task failed", logger.Field("error", err.Error()), logger.Field("payload", string(payloadBuy)))
			continue
		}
		logger.Infow("[CheckSubscription] Send email success",
			logger.Field("taskID", taskInfo.ID), logger.Field("User", sub.UserId),
			logger.Field("Email", method.AuthIdentifier),
		)
	}
	return nil
}

func (l *CheckSubscriptionLogic) sendTrafficNotify(ctx context.Context, subs []int64) error {
	for _, id := range subs {
		sub, err := l.svc.UserModel.FindOneUserSubscribe(ctx, id)
		if err != nil {
			logger.Errorw("[CheckSubscription] FindOneUserSubscribe failed", logger.Field("error", err.Error()))
			continue
		}
		method, err := l.svc.UserModel.FindUserAuthMethodByUserId(ctx, "email", sub.UserId)
		if err != nil {
			logger.Errorw("[CheckSubscription] FindUserAuthMethodByUserId failed", logger.Field("error", err.Error()), logger.Field("user_id", sub.UserId))
			continue
		}
		var taskPayload queue.SendEmailPayload
		taskPayload.Type = queue.EmailTypeTrafficExceed
		taskPayload.Email = method.AuthIdentifier
		taskPayload.Subject = "Subscription Traffic Exceed"
		taskPayload.Content = map[string]interface{}{
			"SiteLogo": l.svc.Config.Site.SiteLogo,
			"SiteName": l.svc.Config.Site.SiteName,
		}
		payloadBuy, err := json.Marshal(taskPayload)
		if err != nil {
			logger.Errorw("[CheckSubscription] Marshal payload failed", logger.Field("error", err.Error()))
			continue
		}
		task := asynq.NewTask(queue.ForthwithSendEmail, payloadBuy, asynq.MaxRetry(3))
		taskInfo, err := l.svc.Queue.Enqueue(task)
		if err != nil {
			logger.Errorw("[CheckSubscription] Enqueue task failed", logger.Field("error", err.Error()), logger.Field("payload", string(payloadBuy)))
			continue
		}
		logger.Infow("[CheckSubscription] Send email success",
			logger.Field("taskID", taskInfo.ID), logger.Field("User", sub.UserId),
			logger.Field("Email", method.AuthIdentifier),
		)
	}
	return nil
}

func (l *CheckSubscriptionLogic) clearServerCache(ctx context.Context, userSubs ...*user.Subscribe) {
	subs := make(map[int64]bool)
	for _, sub := range userSubs {
		if _, ok := subs[sub.SubscribeId]; !ok {
			subs[sub.SubscribeId] = true
		}
	}

	for sub, _ := range subs {
		if err := l.svc.SubscribeModel.ClearCache(ctx, sub); err != nil {
			logger.Errorw("[CheckSubscription] ClearCache failed", logger.Field("error", err.Error()), logger.Field("subscribe_id", sub))
		}
	}
}
