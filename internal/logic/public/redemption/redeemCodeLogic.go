package redemption

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/perfect-panel/server/internal/model/redemption"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/snowflake"
	"github.com/perfect-panel/server/pkg/uuidx"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type RedeemCodeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Redeem code
func NewRedeemCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RedeemCodeLogic {
	return &RedeemCodeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RedeemCodeLogic) RedeemCode(req *types.RedeemCodeRequest) (resp *types.RedeemCodeResponse, err error) {
	// Get user from context
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}

	// Find redemption code by code
	redemptionCode, err := l.svcCtx.RedemptionCodeModel.FindOneByCode(l.ctx, req.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Errorw("[RedeemCode] Redemption code not found", logger.Field("code", req.Code))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "redemption code not found")
		}
		l.Errorw("[RedeemCode] Database Error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find redemption code error: %v", err.Error())
	}

	// Check if redemption code is enabled
	if redemptionCode.Status != 1 {
		l.Errorw("[RedeemCode] Redemption code is disabled",
			logger.Field("code", req.Code),
			logger.Field("status", redemptionCode.Status))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "redemption code is disabled")
	}

	// Check if redemption code has remaining count
	if redemptionCode.TotalCount > 0 && redemptionCode.UsedCount >= redemptionCode.TotalCount {
		l.Errorw("[RedeemCode] Redemption code has been fully used",
			logger.Field("code", req.Code),
			logger.Field("total_count", redemptionCode.TotalCount),
			logger.Field("used_count", redemptionCode.UsedCount))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "redemption code has been fully used")
	}

	// Check if user has already redeemed this code
	userRecords, err := l.svcCtx.RedemptionRecordModel.FindByUserId(l.ctx, u.Id)
	if err != nil {
		l.Errorw("[RedeemCode] Database Error", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find redemption records error: %v", err.Error())
	}
	for _, record := range userRecords {
		if record.RedemptionCodeId == redemptionCode.Id {
			l.Errorw("[RedeemCode] User has already redeemed this code",
				logger.Field("user_id", u.Id),
				logger.Field("code", req.Code))
			return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "you have already redeemed this code")
		}
	}

	// Find subscribe plan from redemption code
	subscribePlan, err := l.svcCtx.SubscribeModel.FindOne(l.ctx, redemptionCode.SubscribePlan)
	if err != nil {
		l.Errorw("[RedeemCode] Subscribe plan not found",
			logger.Field("subscribe_plan", redemptionCode.SubscribePlan),
			logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "subscribe plan not found")
	}

	// Check if subscribe plan is available
	if !*subscribePlan.Sell {
		l.Errorw("[RedeemCode] Subscribe plan is not available",
			logger.Field("subscribe_plan", redemptionCode.SubscribePlan))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.SubscribeNotAvailable), "subscribe plan is not available")
	}

	// Start transaction
	err = l.svcCtx.RedemptionCodeModel.Transaction(l.ctx, func(tx *gorm.DB) error {
		// Find user's existing subscribe for this plan
		var existingSubscribe *user.SubscribeDetails
		userSubscribes, err := l.svcCtx.UserModel.QueryUserSubscribe(l.ctx, u.Id, 0, 1)
		if err == nil {
			for _, us := range userSubscribes {
				if us.SubscribeId == redemptionCode.SubscribePlan {
					existingSubscribe = us
					break
				}
			}
		}

		now := time.Now()

		if existingSubscribe != nil {
			// Extend existing subscribe
			var newExpireTime time.Time
			if existingSubscribe.ExpireTime.After(now) {
				newExpireTime = existingSubscribe.ExpireTime
			} else {
				newExpireTime = now
			}

			// Calculate duration based on redemption code
			duration, err := calculateDuration(redemptionCode.UnitTime, redemptionCode.Quantity)
			if err != nil {
				l.Errorw("[RedeemCode] Calculate duration error", logger.Field("error", err.Error()))
				return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "calculate duration error: %v", err.Error())
			}
			newExpireTime = newExpireTime.Add(duration)

			// Update subscribe
			existingSubscribe.ExpireTime = newExpireTime
			existingSubscribe.Status = 1

			// Add traffic if needed
			if subscribePlan.Traffic > 0 {
				existingSubscribe.Traffic = subscribePlan.Traffic * 1024 * 1024 * 1024
				existingSubscribe.Download = 0
				existingSubscribe.Upload = 0
			}

			err = l.svcCtx.UserModel.UpdateSubscribe(l.ctx, &user.Subscribe{
				Id:         existingSubscribe.Id,
				UserId:     existingSubscribe.UserId,
				ExpireTime: existingSubscribe.ExpireTime,
				Status:     existingSubscribe.Status,
				Traffic:    existingSubscribe.Traffic,
				Download:   existingSubscribe.Download,
				Upload:     existingSubscribe.Upload,
			}, tx)
			if err != nil {
				l.Errorw("[RedeemCode] Update subscribe error", logger.Field("error", err.Error()))
				return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update subscribe error: %v", err.Error())
			}
		} else {
			// Check quota limit before creating new subscribe
			if subscribePlan.Quota > 0 {
				var count int64
				if err := tx.Model(&user.Subscribe{}).Where("user_id = ? AND subscribe_id = ?", u.Id, redemptionCode.SubscribePlan).Count(&count).Error; err != nil {
					l.Errorw("[RedeemCode] Count user subscribe failed", logger.Field("error", err.Error()))
					return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "count user subscribe error: %v", err.Error())
				}
				if count >= subscribePlan.Quota {
					l.Infow("[RedeemCode] Subscribe quota limit exceeded",
						logger.Field("user_id", u.Id),
						logger.Field("subscribe_id", redemptionCode.SubscribePlan),
						logger.Field("quota", subscribePlan.Quota),
						logger.Field("current_count", count),
					)
					return errors.Wrapf(xerr.NewErrCode(xerr.SubscribeQuotaLimit), "subscribe quota limit exceeded")
				}
			}

			// Create new subscribe
			expireTime, traffic, err := calculateSubscribeTimeAndTraffic(redemptionCode.UnitTime, redemptionCode.Quantity, subscribePlan.Traffic)
			if err != nil {
				l.Errorw("[RedeemCode] Calculate subscribe time and traffic error", logger.Field("error", err.Error()))
				return errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "calculate subscribe time and traffic error: %v", err.Error())
			}

			newSubscribe := &user.Subscribe{
				Id:          snowflake.GetID(),
				UserId:      u.Id,
				OrderId:     0,
				SubscribeId: redemptionCode.SubscribePlan,
				StartTime:   now,
				ExpireTime:  expireTime,
				FinishedAt:  nil,
				Traffic:     traffic,
				Download:    0,
				Upload:      0,
				Token:       uuidx.SubscribeToken(fmt.Sprintf("redemption:%d:%d", u.Id, time.Now().UnixMilli())),
				UUID:        uuid.New().String(),
				Status:      1,
			}

			err = l.svcCtx.UserModel.InsertSubscribe(l.ctx, newSubscribe, tx)
			if err != nil {
				l.Errorw("[RedeemCode] Insert subscribe error", logger.Field("error", err.Error()))
				return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "insert subscribe error: %v", err.Error())
			}
		}

		// Increment redemption code used count
		err = l.svcCtx.RedemptionCodeModel.IncrementUsedCount(l.ctx, redemptionCode.Id)
		if err != nil {
			l.Errorw("[RedeemCode] Increment used count error", logger.Field("error", err.Error()))
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "increment used count error: %v", err.Error())
		}

		// Create redemption record
		redemptionRecord := &redemption.RedemptionRecord{
			Id:               snowflake.GetID(),
			RedemptionCodeId: redemptionCode.Id,
			UserId:           u.Id,
			SubscribeId:      redemptionCode.SubscribePlan,
			UnitTime:         redemptionCode.UnitTime,
			Quantity:         redemptionCode.Quantity,
			RedeemedAt:       now,
			CreatedAt:        now,
		}

		err = l.svcCtx.RedemptionRecordModel.Insert(l.ctx, redemptionRecord)
		if err != nil {
			l.Errorw("[RedeemCode] Insert redemption record error", logger.Field("error", err.Error()))
			return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "insert redemption record error: %v", err.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.RedeemCodeResponse{
		Message: "Redemption successful",
	}, nil
}

// calculateDuration calculates time duration based on unit time
func calculateDuration(unitTime string, quantity int64) (time.Duration, error) {
	switch unitTime {
	case "month":
		return time.Duration(quantity*30*24) * time.Hour, nil
	case "quarter":
		return time.Duration(quantity*90*24) * time.Hour, nil
	case "half_year":
		return time.Duration(quantity*180*24) * time.Hour, nil
	case "year":
		return time.Duration(quantity*365*24) * time.Hour, nil
	case "day":
		return time.Duration(quantity*24) * time.Hour, nil
	default:
		return time.Duration(quantity*30*24) * time.Hour, nil
	}
}

// calculateSubscribeTimeAndTraffic calculates expire time and traffic based on subscribe plan
func calculateSubscribeTimeAndTraffic(unitTime string, quantity int64, traffic int64) (time.Time, int64, error) {
	duration, err := calculateDuration(unitTime, quantity)
	if err != nil {
		return time.Time{}, 0, err
	}

	expireTime := time.Now().Add(duration)
	trafficBytes := int64(0)
	if traffic > 0 {
		trafficBytes = traffic * 1024 * 1024 * 1024
	}

	return expireTime, trafficBytes, nil
}
