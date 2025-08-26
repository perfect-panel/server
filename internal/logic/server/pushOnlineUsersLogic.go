package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/perfect-panel/server/internal/model/cache"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type PushOnlineUsersLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPushOnlineUsersLogic Push online users
func NewPushOnlineUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PushOnlineUsersLogic {
	return &PushOnlineUsersLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PushOnlineUsersLogic) PushOnlineUsers(req *types.OnlineUsersRequest) error {
	// 验证请求数据
	if req.ServerId <= 0 || len(req.Users) == 0 {
		return errors.New("invalid request parameters")
	}

	// 验证用户数据
	for _, user := range req.Users {
		if user.SID <= 0 || user.IP == "" {
			return fmt.Errorf("invalid user data: uid=%d, ip=%s", user.SID, user.IP)
		}
	}

	// Find server info
	_, err := l.svcCtx.NodeModel.FindOneServer(l.ctx, req.ServerId)
	if err != nil {
		l.Errorw("[PushOnlineUsers] FindOne error", logger.Field("error", err))
		return fmt.Errorf("server not found: %w", err)
	}

	userOnlineIp := make([]cache.NodeOnlineUser, 0)
	for _, user := range req.Users {
		userOnlineIp = append(userOnlineIp, cache.NodeOnlineUser{
			SID: user.SID,
			IP:  user.IP,
		})
	}
	err = l.svcCtx.NodeCache.AddOnlineUserIP(l.ctx, userOnlineIp)
	if err != nil {
		l.Errorw("[PushOnlineUsers] cache operation error", logger.Field("error", err))
		return err
	}

	err = l.svcCtx.NodeCache.UpdateNodeOnlineUser(l.ctx, req.ServerId, userOnlineIp)

	if err != nil {
		l.Errorw("[PushOnlineUsers] cache operation error", logger.Field("error", err))
		return err
	}

	return nil
}
