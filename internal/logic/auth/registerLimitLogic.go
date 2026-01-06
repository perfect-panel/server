package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/svc"
	"go.uber.org/zap"
)

func registerIpLimit(svcCtx *svc.ServiceContext, ctx context.Context, registerIp, authType, account string) (isOk bool) {
	if !svcCtx.Config.Register.EnableIpRegisterLimit {
		return true
	}
	cacheKey := fmt.Sprintf("%s%s:*", config.RegisterIpKeyPrefix, registerIp)
	var cacheKeys []string
	var cursor uint64
	for {
		keys, newCursor, err := svcCtx.Redis.Scan(ctx, 0, cacheKey, 100).Result()
		if err != nil {
			zap.S().Errorf("[registerIpLimit] Err: %v", err)
			return true
		}
		if len(keys) > 0 {
			cacheKeys = append(cacheKeys, keys...)
		}
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	defer func() {
		key := fmt.Sprintf("%s%s:%s:%s", config.RegisterIpKeyPrefix, registerIp, authType, account)
		if err := svcCtx.Redis.Set(ctx, key, account, time.Minute*time.Duration(svcCtx.Config.Register.IpRegisterLimitDuration)).Err(); err != nil {
			zap.S().Errorf("[registerIpLimit] Set Err: %v", err)
		}
	}()
	if len(cacheKeys) < int(svcCtx.Config.Register.IpRegisterLimit) {
		return true
	}
	return false
}
