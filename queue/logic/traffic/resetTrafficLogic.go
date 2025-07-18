package traffic

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/server/internal/model/subscribe"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/queue/types"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ResetTrafficLogic handles traffic reset logic for different subscription cycles
// Supports three reset modes:
// - reset_cycle = 1: Reset on 1st of every month
// - reset_cycle = 2: Reset monthly based on subscription start date
// - reset_cycle = 3: Reset yearly based on subscription start date
type ResetTrafficLogic struct {
	svc *svc.ServiceContext
}

// Cache and retry configuration constants
const (
	maxRetryAttempts = 3
	retryDelay       = 30 * time.Minute
	lockTimeout      = 5 * time.Minute
)

// Cache keys
var (
	cacheKey      = "reset_traffic_cache"
	retryCountKey = "reset_traffic_retry_count"
	lockKey       = "reset_traffic_lock"
)

// resetTrafficCache stores the last reset time to prevent duplicate processing
type resetTrafficCache struct {
	LastResetTime time.Time
}

func NewResetTrafficLogic(svc *svc.ServiceContext) *ResetTrafficLogic {
	return &ResetTrafficLogic{
		svc: svc,
	}
}

// ProcessTask executes the traffic reset task for all subscription types with enhanced retry mechanism
func (l *ResetTrafficLogic) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	var err error
	startTime := time.Now()

	// Get current retry count
	retryCount := l.getRetryCount(ctx)
	logger.Infow("[ResetTraffic] Starting task execution",
		logger.Field("retryCount", retryCount),
		logger.Field("startTime", startTime))

	// Acquire distributed lock to prevent duplicate execution
	lockAcquired := l.acquireLock(ctx)
	if !lockAcquired {
		logger.Infow("[ResetTraffic] Another task is already running, skipping execution")
		return nil
	}
	defer l.releaseLock(ctx)

	defer func() {
		if err != nil {
			// Check if error is retryable and within retry limit
			if l.isRetryableError(err) && retryCount < maxRetryAttempts {
				// Increment retry count
				l.setRetryCount(ctx, retryCount+1)

				// Schedule retry with delay
				task := asynq.NewTask(types.SchedulerResetTraffic, nil)
				_, retryErr := l.svc.Queue.Enqueue(task, asynq.ProcessIn(retryDelay))
				if retryErr != nil {
					logger.Errorw("[ResetTraffic] Failed to enqueue retry task",
						logger.Field("error", retryErr.Error()),
						logger.Field("retryCount", retryCount))
				} else {
					logger.Infow("[ResetTraffic] Task failed, retrying in 30 minutes",
						logger.Field("error", err.Error()),
						logger.Field("retryCount", retryCount+1),
						logger.Field("maxRetryAttempts", maxRetryAttempts))
				}
			} else {
				// Max retries reached or non-retryable error
				if retryCount >= maxRetryAttempts {
					logger.Errorw("[ResetTraffic] Max retry attempts reached, giving up",
						logger.Field("retryCount", retryCount),
						logger.Field("maxRetryAttempts", maxRetryAttempts),
						logger.Field("error", err.Error()))
				} else {
					logger.Errorw("[ResetTraffic] Non-retryable error, not retrying",
						logger.Field("error", err.Error()),
						logger.Field("retryCount", retryCount))
				}
				// Reset retry count for next scheduled task
				l.clearRetryCount(ctx)
			}
		} else {
			// Task completed successfully, reset retry count
			l.clearRetryCount(ctx)
			logger.Infow("[ResetTraffic] Task completed successfully",
				logger.Field("processingTime", time.Since(startTime)),
				logger.Field("retryCount", retryCount))
		}
	}()

	// Load last reset time from cache
	var cache resetTrafficCache
	err = l.svc.Redis.Get(ctx, cacheKey).Scan(&cache)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Errorw("[ResetTraffic] Failed to get cache", logger.Field("error", err.Error()))
		}
		// Set default value if cache not found
		cache = resetTrafficCache{
			LastResetTime: time.Now().Add(-10 * time.Minute),
		}
		logger.Infow("[ResetTraffic] Using default cache value", logger.Field("lastResetTime", cache.LastResetTime))
	} else {
		logger.Infow("[ResetTraffic] Cache loaded successfully", logger.Field("lastResetTime", cache.LastResetTime))
	}

	// Execute reset operations in order: yearly -> monthly (1st) -> monthly (cycle)
	err = l.resetYear(ctx)
	if err != nil {
		logger.Errorw("[ResetTraffic] Yearly reset failed", logger.Field("error", err.Error()))
		return err
	}

	err = l.reset1st(ctx, cache)
	if err != nil {
		logger.Errorw("[ResetTraffic] Monthly 1st reset failed", logger.Field("error", err.Error()))
		return err
	}

	err = l.resetMonth(ctx)
	if err != nil {
		logger.Errorw("[ResetTraffic] Monthly cycle reset failed", logger.Field("error", err.Error()))
		return err
	}

	// Update cache with current time after successful processing
	updatedCache := resetTrafficCache{
		LastResetTime: startTime,
	}
	cacheErr := l.svc.Redis.Set(ctx, cacheKey, updatedCache, 0).Err()
	if cacheErr != nil {
		logger.Errorw("[ResetTraffic] Failed to update cache", logger.Field("error", cacheErr.Error()))
		// Don't return error here as the main task completed successfully
	} else {
		logger.Infow("[ResetTraffic] Cache updated successfully", logger.Field("newLastResetTime", startTime))
	}

	return nil
}

// resetMonth handles monthly cycle reset based on subscription start date
// reset_cycle = 2: Reset monthly based on subscription start date
func (l *ResetTrafficLogic) resetMonth(ctx context.Context) error {
	now := time.Now()

	err := l.svc.UserModel.Transaction(ctx, func(db *gorm.DB) error {
		// Get all subscriptions that reset monthly based on start date
		var resetMonthSubIds []int64
		err := db.Model(&subscribe.Subscribe{}).Select("`id`").Where("`reset_cycle` = ?", 2).Find(&resetMonthSubIds).Error
		if err != nil {
			logger.Errorw("[ResetTraffic] Failed to query monthly subscriptions", logger.Field("error", err.Error()))
			return err
		}

		if len(resetMonthSubIds) == 0 {
			logger.Infow("[ResetTraffic] No monthly cycle subscriptions found")
			return nil
		}

		// Query users for monthly reset based on subscription start date cycle
		var monthlyResetUsers []int64

		// Check if today is the last day of current month
		nextMonth := now.AddDate(0, 1, 0)
		isLastDayOfMonth := nextMonth.Month() != now.Month()

		query := db.Model(&user.Subscribe{}).Select("`id`").
			Where("`subscribe_id` IN ?", resetMonthSubIds).
			Where("`status` = ?", 1).                                                                          // Only active subscriptions
			Where("PERIOD_DIFF(DATE_FORMAT(CURDATE(), '%Y%m'), DATE_FORMAT(start_time, '%Y%m')) > 0").         // At least one month passed
			Where("MOD(PERIOD_DIFF(DATE_FORMAT(CURDATE(), '%Y%m'), DATE_FORMAT(start_time, '%Y%m')), 1) = 0"). // Monthly cycle
			Where("DATE(start_time) < CURDATE()")                                                              // Only reset subscriptions that have started

		if isLastDayOfMonth {
			// Last day of month: handle subscription start dates >= today
			query = query.Where("DAY(start_time) >= ?", now.Day())
		} else {
			// Normal case: exact day match
			query = query.Where("DAY(start_time) = ?", now.Day())
		}

		err = query.Find(&monthlyResetUsers).Error
		if err != nil {
			logger.Errorw("[ResetTraffic] Failed to query monthly reset users", logger.Field("error", err.Error()))
			return err
		}

		if len(monthlyResetUsers) > 0 {
			logger.Infow("[ResetTraffic] Found users for monthly reset",
				logger.Field("count", len(monthlyResetUsers)),
				logger.Field("userIds", monthlyResetUsers))

			err = db.Model(&user.Subscribe{}).Where("`id` IN ?", monthlyResetUsers).
				Updates(map[string]interface{}{
					"upload":   0,
					"download": 0,
				}).Error
			if err != nil {
				logger.Errorw("[ResetTraffic] Failed to update monthly reset users", logger.Field("error", err.Error()))
				return err
			}

			logger.Infow("[ResetTraffic] Monthly reset completed", logger.Field("count", len(monthlyResetUsers)))
		} else {
			logger.Infow("[ResetTraffic] No users found for monthly reset")
		}

		return nil
	})
	if err != nil {
		logger.Errorw("[ResetTraffic] Monthly reset transaction failed", logger.Field("error", err.Error()))
		return err
	}

	logger.Infow("[ResetTraffic] Monthly reset process completed")
	return nil
}

// reset1st handles reset on 1st of every month
// reset_cycle = 1: Reset on 1st of every month
func (l *ResetTrafficLogic) reset1st(ctx context.Context, cache resetTrafficCache) error {
	now := time.Now()

	// Check if we already reset this month using cache
	if cache.LastResetTime.Year() == now.Year() && cache.LastResetTime.Month() == now.Month() {
		logger.Infow("[ResetTraffic] Already reset this month, skipping 1st reset",
			logger.Field("lastResetTime", cache.LastResetTime),
			logger.Field("currentTime", now))
		return nil
	}

	// Only reset if it's the 1st day of the month
	if now.Day() != 1 {
		logger.Infow("[ResetTraffic] Not 1st day of month, skipping 1st reset", logger.Field("currentDay", now.Day()))
		return nil
	}

	err := l.svc.UserModel.Transaction(ctx, func(db *gorm.DB) error {
		// Get all subscriptions that reset on 1st of month
		var reset1stSubIds []int64
		err := db.Model(&subscribe.Subscribe{}).Select("`id`").Where("`reset_cycle` = ?", 1).Find(&reset1stSubIds).Error
		if err != nil {
			logger.Errorw("[ResetTraffic] Failed to query 1st reset subscriptions", logger.Field("error", err.Error()))
			return err
		}

		if len(reset1stSubIds) == 0 {
			logger.Infow("[ResetTraffic] No 1st reset subscriptions found")
			return nil
		}

		// Get all active users with these subscriptions
		var users1stReset []int64
		err = db.Model(&user.Subscribe{}).Select("`id`").
			Where("`subscribe_id` IN ?", reset1stSubIds).
			Where("`status` = ?", 1). // Only active subscriptions
			Find(&users1stReset).Error
		if err != nil {
			logger.Errorw("[ResetTraffic] Failed to query 1st reset users", logger.Field("error", err.Error()))
			return err
		}

		if len(users1stReset) > 0 {
			logger.Infow("[ResetTraffic] Found users for 1st reset",
				logger.Field("count", len(users1stReset)),
				logger.Field("userIds", users1stReset))

			// Reset upload and download traffic to zero
			err = db.Model(&user.Subscribe{}).Where("`id` IN ?", users1stReset).
				Updates(map[string]interface{}{
					"upload":   0,
					"download": 0,
				}).Error
			if err != nil {
				logger.Errorw("[ResetTraffic] Failed to update 1st reset users", logger.Field("error", err.Error()))
				return err
			}

			logger.Infow("[ResetTraffic] 1st reset completed", logger.Field("count", len(users1stReset)))
		} else {
			logger.Infow("[ResetTraffic] No users found for 1st reset")
		}

		return nil
	})

	if err != nil {
		logger.Errorw("[ResetTraffic] 1st reset transaction failed", logger.Field("error", err.Error()))
		return err
	}

	logger.Infow("[ResetTraffic] 1st reset process completed")
	return nil
}

// resetYear handles yearly reset based on subscription start date anniversary
// reset_cycle = 3: Reset yearly based on subscription start date
func (l *ResetTrafficLogic) resetYear(ctx context.Context) error {
	now := time.Now()

	err := l.svc.UserModel.Transaction(ctx, func(db *gorm.DB) error {
		// Get all subscriptions that reset yearly
		var resetYearSubIds []int64
		err := db.Model(&subscribe.Subscribe{}).Select("`id`").Where("`reset_cycle` = ?", 3).Find(&resetYearSubIds).Error
		if err != nil {
			logger.Errorw("[ResetTraffic] Failed to query yearly subscriptions", logger.Field("error", err.Error()))
			return err
		}

		if len(resetYearSubIds) == 0 {
			logger.Infow("[ResetTraffic] No yearly reset subscriptions found")
			return nil
		}

		// Query users for yearly reset based on subscription start date anniversary
		var usersYearReset []int64

		// Check if today is the last day of current month
		nextMonth := now.AddDate(0, 1, 0)
		isLastDayOfMonth := nextMonth.Month() != now.Month()

		// Check if today is February 28th (handle leap year case)
		isLeapYearCase := now.Month() == 2 && now.Day() == 28

		query := db.Model(&user.Subscribe{}).Select("`id`").
			Where("`subscribe_id` IN ?", resetYearSubIds).
			Where("MONTH(start_time) = ?", now.Month()).                    // Same month
			Where("`status` = ?", 1).                                       // Only active subscriptions
			Where("TIMESTAMPDIFF(YEAR, DATE(start_time), CURDATE()) >= 1"). // At least 1 year passed
			Where("DATE(start_time) < CURDATE()")                           // Only reset subscriptions that have started

		if isLeapYearCase {
			// February 28th: handle both Feb 28 and Feb 29 subscriptions
			query = query.Where("DAY(start_time) IN (28, 29)")
		} else if isLastDayOfMonth {
			// Last day of month: handle subscription start dates >= today
			query = query.Where("DAY(start_time) >= ?", now.Day())
		} else {
			// Normal case: exact day match
			query = query.Where("DAY(start_time) = ?", now.Day())
		}

		err = query.Find(&usersYearReset).Error
		if err != nil {
			logger.Errorw("[ResetTraffic] Query yearly reset users failed", logger.Field("error", err.Error()))
			return err
		}

		if len(usersYearReset) > 0 {
			logger.Infow("[ResetTraffic] Found users for yearly reset",
				logger.Field("count", len(usersYearReset)),
				logger.Field("userIds", usersYearReset))

			// Reset upload and download traffic to zero
			err = db.Model(&user.Subscribe{}).Where("`id` IN ?", usersYearReset).
				Updates(map[string]interface{}{
					"upload":   0,
					"download": 0,
				}).Error
			if err != nil {
				logger.Errorw("[ResetTraffic] Failed to update yearly reset users", logger.Field("error", err.Error()))
				return err
			}

			logger.Infow("[ResetTraffic] Yearly reset completed", logger.Field("count", len(usersYearReset)))
		} else {
			logger.Infow("[ResetTraffic] No users found for yearly reset")
		}

		return nil
	})

	if err != nil {
		logger.Errorw("[ResetTraffic] Yearly reset transaction failed", logger.Field("error", err.Error()))
		return err
	}

	logger.Infow("[ResetTraffic] Yearly reset process completed")
	return nil
}

// getRetryCount retrieves the current retry count from Redis
func (l *ResetTrafficLogic) getRetryCount(ctx context.Context) int {
	countStr, err := l.svc.Redis.Get(ctx, retryCountKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0 // No retry count found, start with 0
		}
		logger.Errorw("[ResetTraffic] Failed to get retry count", logger.Field("error", err.Error()))
		return 0
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		logger.Errorw("[ResetTraffic] Invalid retry count format", logger.Field("value", countStr))
		return 0
	}

	return count
}

// setRetryCount sets the retry count in Redis
func (l *ResetTrafficLogic) setRetryCount(ctx context.Context, count int) {
	err := l.svc.Redis.Set(ctx, retryCountKey, count, 24*time.Hour).Err()
	if err != nil {
		logger.Errorw("[ResetTraffic] Failed to set retry count",
			logger.Field("count", count),
			logger.Field("error", err.Error()))
	}
}

// clearRetryCount removes the retry count from Redis
func (l *ResetTrafficLogic) clearRetryCount(ctx context.Context) {
	err := l.svc.Redis.Del(ctx, retryCountKey).Err()
	if err != nil {
		logger.Errorw("[ResetTraffic] Failed to clear retry count", logger.Field("error", err.Error()))
	}
}

// acquireLock attempts to acquire a distributed lock
func (l *ResetTrafficLogic) acquireLock(ctx context.Context) bool {
	result := l.svc.Redis.SetNX(ctx, lockKey, "locked", lockTimeout)
	acquired, err := result.Result()
	if err != nil {
		logger.Errorw("[ResetTraffic] Failed to acquire lock", logger.Field("error", err.Error()))
		return false
	}

	if acquired {
		logger.Infow("[ResetTraffic] Lock acquired successfully")
	} else {
		logger.Infow("[ResetTraffic] Lock already exists, another task is running")
	}

	return acquired
}

// releaseLock releases the distributed lock
func (l *ResetTrafficLogic) releaseLock(ctx context.Context) {
	err := l.svc.Redis.Del(ctx, lockKey).Err()
	if err != nil {
		logger.Errorw("[ResetTraffic] Failed to release lock", logger.Field("error", err.Error()))
	} else {
		logger.Infow("[ResetTraffic] Lock released successfully")
	}
}

// isRetryableError determines if an error is retryable
func (l *ResetTrafficLogic) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorMessage := strings.ToLower(err.Error())

	// Network and connection errors (retryable)
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"network",
		"timeout",
		"dial",
		"context deadline exceeded",
		"temporary failure",
		"server error",
		"service unavailable",
		"internal server error",
		"database is locked",
		"too many connections",
		"deadlock",
		"lock wait timeout",
	}

	// Database constraint errors (non-retryable)
	nonRetryableErrors := []string{
		"foreign key constraint",
		"unique constraint",
		"check constraint",
		"not null constraint",
		"invalid input syntax",
		"column does not exist",
		"table does not exist",
		"permission denied",
		"access denied",
		"authentication failed",
		"invalid credentials",
	}

	// Check for non-retryable errors first
	for _, nonRetryable := range nonRetryableErrors {
		if strings.Contains(errorMessage, nonRetryable) {
			logger.Infow("[ResetTraffic] Non-retryable error detected",
				logger.Field("error", err.Error()),
				logger.Field("pattern", nonRetryable))
			return false
		}
	}

	// Check for retryable errors
	for _, retryable := range retryableErrors {
		if strings.Contains(errorMessage, retryable) {
			logger.Infow("[ResetTraffic] Retryable error detected",
				logger.Field("error", err.Error()),
				logger.Field("pattern", retryable))
			return true
		}
	}

	// Default: treat unknown errors as retryable, but log for analysis
	logger.Infow("[ResetTraffic] Unknown error type, treating as retryable",
		logger.Field("error", err.Error()))
	return true
}
