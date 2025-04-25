package server

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/perfect-panel/ppanel-server/internal/model/cache"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/internal/types"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/tool"
	task "github.com/perfect-panel/ppanel-server/queue/types"
	"github.com/pkg/errors"
)

//goland:noinspection GoNameStartsWithPackageName
type ServerPushUserTrafficLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewServerPushUserTrafficLogic Push user Traffic
func NewServerPushUserTrafficLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ServerPushUserTrafficLogic {
	return &ServerPushUserTrafficLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ServerPushUserTrafficLogic) ServerPushUserTraffic(req *types.ServerPushUserTrafficRequest) error {
	// Find server info
	serverInfo, err := l.svcCtx.ServerModel.FindOne(l.ctx, req.ServerId)
	if err != nil {
		l.Errorw("[PushOnlineUsers] FindOne error", logger.Field("error", err))
		return errors.New("server not found")
	}

	// Create traffic task
	var request task.TrafficStatistics
	var userTraffic []cache.UserTraffic
	request.ServerId = serverInfo.Id
	tool.DeepCopy(&request.Logs, req.Traffic)
	tool.DeepCopy(&userTraffic, req.Traffic)

	// update today traffic rank
	err = l.svcCtx.NodeCache.AddNodeTodayTraffic(l.ctx, serverInfo.Id, userTraffic)
	if err != nil {
		l.Errorw("[ServerPushUserTraffic] AddNodeTodayTraffic error", logger.Field("error", err))
		return errors.New("add node today traffic error")
	}
	for _, user := range req.Traffic {
		if err = l.svcCtx.NodeCache.AddUserTodayTraffic(l.ctx, user.SID, user.Upload, user.Download); err != nil {
			l.Errorw("[ServerPushUserTraffic] AddUserTodayTraffic error", logger.Field("error", err))
			continue
		}
	}
	// Push traffic task
	val, _ := json.Marshal(request)
	t := asynq.NewTask(task.ForthwithTrafficStatistics, val, asynq.MaxRetry(3))
	info, err := l.svcCtx.Queue.EnqueueContext(l.ctx, t)
	if err != nil {
		l.Errorw("[ServerPushUserTraffic] Push traffic task error", logger.Field("error", err.Error()), logger.Field("task", t))
	} else {
		l.Infow("[ServerPushUserTraffic] Push traffic task success", logger.Field("task", t), logger.Field("info", info))
	}
	return nil
}
