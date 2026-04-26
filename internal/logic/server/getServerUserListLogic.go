package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/model/subscribe"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/uuidx"
	"github.com/perfect-panel/server/pkg/xerr"
)

// V4.3 节点 user list 重构(决策 17 / 31 / 32 / 40):
// - 输出粒度由 user_subscribe → user_subscribe_device(每设备一行,Id = device.id)
// - 跨协议 limiter 用 user_subscribe_id 作 key(决策 32)
// - SS 协议密码 = sha256(uuid)[:16] 派生(决策 17)
// - subscribe.cut_off_at < now → 整组 device 不下发,实现"24h 后断网"(决策 40)
// - subscribe.throttled_at != nil → 把 speed_limit 压到 1 Mbps(决策 31)
// - device.status == 0(停用)→ 不下发

const (
	throttledSpeedLimit = 1 * 1024 * 1024 // 1 MiB/s ≈ 1 Mbps(MiB 进制,决策 31/30)
	deviceLimitPerSlot  = 1               // 每 UUID 内置 device_limit=1,防 URL 共享(决策 32)
)

type GetServerUserListLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewGetServerUserListLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *GetServerUserListLogic {
	return &GetServerUserListLogic{
		Logger: logger.WithContext(ctx.Request.Context()),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerUserListLogic) GetServerUserList(req *types.GetServerUserListRequest) (resp *types.GetServerUserListResponse, err error) {
	cacheKey := fmt.Sprintf("%s%d", node.ServerUserListCacheKey, req.ServerId)
	cache, _ := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
	if cache != "" {
		etag := tool.GenerateETag([]byte(cache))
		if match := l.ctx.GetHeader("If-None-Match"); match == etag {
			return nil, xerr.StatusNotModified
		}
		l.ctx.Header("ETag", etag)
		resp = &types.GetServerUserListResponse{}
		if err = json.Unmarshal([]byte(cache), resp); err != nil {
			l.Errorw("[ServerUserListCacheKey] json unmarshal error", logger.Field("error", err.Error()))
			return nil, err
		}
		return resp, nil
	}

	server, err := l.svcCtx.NodeModel.FindOneServer(l.ctx, req.ServerId)
	if err != nil {
		return nil, err
	}

	_, nodes, err := l.svcCtx.NodeModel.FilterNodeList(l.ctx, &node.FilterNodeParams{
		Page:     1,
		Size:     1000,
		ServerId: []int64{server.Id},
		Protocol: req.Protocol,
	})
	if err != nil {
		l.Errorw("FilterNodeList error", logger.Field("error", err.Error()))
		return nil, err
	}
	var nodeTag []string
	var nodeIds []int64
	for _, n := range nodes {
		nodeIds = append(nodeIds, n.Id)
		if n.Tags != "" {
			nodeTag = append(nodeTag, strings.Split(n.Tags, ",")...)
		}
	}
	_, subs, err := l.svcCtx.SubscribeModel.FilterList(l.ctx, &subscribe.FilterParams{
		Page: 1,
		Size: 9999,
		Node: nodeIds,
		Tags: nodeTag,
	})
	if err != nil {
		l.Errorw("Subscribe FilterList error", logger.Field("error", err.Error()))
		return nil, err
	}
	if len(subs) == 0 {
		return placeholderResp(), nil
	}

	now := time.Now()
	users := make([]types.ServerUser, 0, 64)
	for _, plan := range subs {
		userSubs, err := l.svcCtx.UserModel.FindUsersSubscribeBySubscribeId(l.ctx, plan.Id)
		if err != nil {
			l.Errorw("FindUsersSubscribeBySubscribeId error",
				logger.Field("error", err.Error()),
				logger.Field("subscribe_id", plan.Id))
			continue
		}
		for _, sub := range userSubs {
			// 决策 40:24h 后整组断网,直接跳过该 subscribe 的所有 device
			if sub.CutOffAt != nil && now.After(*sub.CutOffAt) {
				continue
			}

			// 决策 31:超额限速到 1 Mbps,优先级链取 min(节点级, 套餐级, 1 Mbps)
			speedLimit := plan.SpeedLimit
			if sub.ThrottledAt != nil {
				speedLimit = throttledSpeedLimit
			}

			// 拉本订阅的所有设备槽
			devices, err := l.svcCtx.UserModel.QuerySubscribeDevices(l.ctx, sub.Id)
			if err != nil {
				l.Errorw("QuerySubscribeDevices error",
					logger.Field("error", err.Error()),
					logger.Field("user_subscribe_id", sub.Id))
				continue
			}
			// 兼容老数据:无设备槽的 subscribe(legacy 单 UUID 模式)→ 退化下发 1 行(沿用旧 token/uuid)
			if len(devices) == 0 && sub.UUID != "" {
				users = append(users, types.ServerUser{
					Id:              sub.Id,
					UUID:            sub.UUID,
					SpeedLimit:      speedLimit,
					DeviceLimit:     plan.DeviceLimit,
					UserSubscribeId: sub.Id,
					Password:        tool.DerivePasswordFromUUID(sub.UUID),
				})
				continue
			}
			for _, d := range devices {
				if d.Status == 0 {
					continue // 停用槽不下发
				}
				users = append(users, types.ServerUser{
					Id:              d.Id,
					UUID:            d.UUID,
					SpeedLimit:      speedLimit,
					DeviceLimit:     deviceLimitPerSlot,
					UserSubscribeId: sub.Id,
					Password:        tool.DerivePasswordFromUUID(d.UUID),
				})
			}
		}
	}
	if len(users) == 0 {
		return placeholderResp(), nil
	}
	resp = &types.GetServerUserListResponse{Users: users}
	val, _ := json.Marshal(resp)
	etag := tool.GenerateETag(val)
	l.ctx.Header("ETag", etag)
	if err = l.svcCtx.Redis.Set(l.ctx, cacheKey, string(val), -1).Err(); err != nil {
		l.Errorw("[ServerUserListCacheKey] redis set error", logger.Field("error", err.Error()))
	}
	if match := l.ctx.GetHeader("If-None-Match"); match == etag {
		return nil, xerr.StatusNotModified
	}
	return resp, nil
}

// placeholderResp returns 1-element list so the node never sees empty users
// (some node implementations treat empty as "config error" rather than "no users").
func placeholderResp() *types.GetServerUserListResponse {
	return &types.GetServerUserListResponse{
		Users: []types.ServerUser{{Id: 1, UUID: uuidx.NewUUID().String()}},
	}
}
