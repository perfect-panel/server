package server

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/uuidx"
	"github.com/perfect-panel/server/pkg/xerr"
)

type GetServerUserListLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

// Get user list
func NewGetServerUserListLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *GetServerUserListLogic {
	return &GetServerUserListLogic{
		Logger: logger.WithContext(ctx.Request.Context()),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerUserListLogic) GetServerUserList(req *types.GetServerUserListRequest) (resp *types.GetServerUserListResponse, err error) {
	cacheKey := fmt.Sprintf("%s%d", config.ServerUserListCacheKey, req.ServerId)
	cache, err := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
	if err == nil {
		if cache != "" {
			etag := tool.GenerateETag([]byte(cache))
			resp := &types.GetServerUserListResponse{}
			//  Check If-None-Match header
			if match := l.ctx.GetHeader("If-None-Match"); match == etag {
				return nil, xerr.StatusNotModified
			}
			l.ctx.Header("ETag", etag)
			err = json.Unmarshal([]byte(cache), resp)
			if err != nil {
				l.Errorw("[ServerUserListCacheKey] json unmarshal error", logger.Field("error", err.Error()))
				return nil, err
			}
			return resp, nil
		}
	}
	server, err := l.svcCtx.ServerModel.FindOne(l.ctx, req.ServerId)
	if err != nil {
		return nil, err
	}
	subs, err := l.svcCtx.SubscribeModel.QuerySubscribeIdsByServerIdAndServerGroupId(l.ctx, server.Id, server.GroupId)
	if err != nil {
		l.Errorw("QuerySubscribeIdsByServerIdAndServerGroupId error", logger.Field("error", err.Error()))
		return nil, err
	}
	if len(subs) == 0 {
		return &types.GetServerUserListResponse{
			Users: []types.ServerUser{
				{
					Id:   1,
					UUID: uuidx.NewUUID().String(),
				},
			},
		}, nil
	}
	users := make([]types.ServerUser, 0)
	for _, sub := range subs {
		data, err := l.svcCtx.UserModel.FindUsersSubscribeBySubscribeId(l.ctx, sub.Id)
		if err != nil {
			return nil, err
		}
		for _, datum := range data {
			speedLimit := server.SpeedLimit
			if (int(sub.SpeedLimit) < server.SpeedLimit && sub.SpeedLimit != 0) ||
				(int(sub.SpeedLimit) > server.SpeedLimit && sub.SpeedLimit == 0) {
				speedLimit = int(sub.SpeedLimit)
			}

			users = append(users, types.ServerUser{
				Id:          datum.Id,
				UUID:        datum.UUID,
				SpeedLimit:  int64(speedLimit),
				DeviceLimit: sub.DeviceLimit,
			})
		}
	}
	if len(users) == 0 {
		users = append(users, types.ServerUser{
			Id:   1,
			UUID: uuidx.NewUUID().String(),
		})
	}
	resp = &types.GetServerUserListResponse{
		Users: users,
	}
	val, _ := json.Marshal(resp)
	etag := tool.GenerateETag(val)
	l.ctx.Header("ETag", etag)
	err = l.svcCtx.Redis.Set(l.ctx, cacheKey, string(val), -1).Err()
	if err != nil {
		l.Errorw("[ServerUserListCacheKey] redis set error", logger.Field("error", err.Error()))
	}
	return resp, nil
}
