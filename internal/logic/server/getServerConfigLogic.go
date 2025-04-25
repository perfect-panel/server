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
	"github.com/perfect-panel/server/pkg/xerr"
)

type GetServerConfigLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

// Get server config
func NewGetServerConfigLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *GetServerConfigLogic {
	return &GetServerConfigLogic{
		Logger: logger.WithContext(ctx.Request.Context()),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerConfigLogic) GetServerConfig(req *types.GetServerConfigRequest) (resp *types.GetServerConfigResponse, err error) {
	cacheKey := fmt.Sprintf("%s%d", config.ServerConfigCacheKey, req.ServerId)
	cache, err := l.svcCtx.Redis.Get(l.ctx, cacheKey).Result()
	if err == nil {
		if cache != "" {
			etag := tool.GenerateETag([]byte(cache))
			//  Check If-None-Match header
			match := l.ctx.GetHeader("If-None-Match")
			if match == etag {
				return nil, xerr.StatusNotModified
			}
			l.ctx.Header("ETag", etag)
			resp := &types.GetServerConfigResponse{}
			err = json.Unmarshal([]byte(cache), resp)
			if err != nil {
				l.Errorw("[ServerConfigCacheKey] json unmarshal error", logger.Field("error", err.Error()))
				return nil, err
			}
			return resp, nil
		}
	}
	nodeInfo, err := l.svcCtx.ServerModel.FindOne(l.ctx, req.ServerId)
	if err != nil {
		l.Errorw("[GetServerConfig] FindOne error", logger.Field("error", err.Error()))
		return nil, err
	}
	cfg := make(map[string]interface{})
	err = json.Unmarshal([]byte(nodeInfo.Config), &cfg)
	if err != nil {
		l.Errorw("[GetServerConfig] json unmarshal error", logger.Field("error", err.Error()))
		return nil, err
	}
	resp = &types.GetServerConfigResponse{
		Basic: types.ServerBasic{
			PullInterval: l.svcCtx.Config.Node.NodePullInterval,
			PushInterval: l.svcCtx.Config.Node.NodePushInterval,
		},
		Protocol: nodeInfo.Protocol,
		Config:   cfg,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		l.Errorw("[GetServerConfig] json marshal error", logger.Field("error", err.Error()))
		return nil, err
	}
	etag := tool.GenerateETag(data)
	l.ctx.Header("ETag", etag)
	if err = l.svcCtx.Redis.Set(l.ctx, cacheKey, data, -1).Err(); err != nil {
		l.Errorw("[GetServerConfig] redis set error", logger.Field("error", err.Error()))
	}
	return resp, nil
}
