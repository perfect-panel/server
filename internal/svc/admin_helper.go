package svc

// Shared helpers for admin modules (sitecontent / audit / device / server).
// Placed in the svc package so per-module logic packages don't import each other.

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/constant"
)

// InvalidateNodeCache wipes Redis keys matching `server:user:*` so node-side
// caches refetch immediately after admin device ops (disable/enable/reset).
func InvalidateNodeCache(ctx *gin.Context, svcCtx *ServiceContext) {
	pattern := "server:user:*"
	iter := svcCtx.Redis.Scan(ctx, 0, pattern, 100).Iterator()
	keys := make([]string, 0, 16)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		_ = svcCtx.Redis.Del(ctx, keys...).Err()
	}
}

// AdminIdFromCtx pulls the operating admin user.id out of the gin/auth context.
// Returns (0, false) when no authenticated user is present.
func AdminIdFromCtx(ctx context.Context) (int64, bool) {
	if u, ok := ctx.Value(constant.CtxKeyUser).(*user.User); ok {
		return u.Id, true
	}
	return 0, false
}
