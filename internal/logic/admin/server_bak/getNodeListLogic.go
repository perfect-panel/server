package server_bak

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/perfect-panel/server/internal/model/server"

	"github.com/redis/go-redis/v9"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetNodeListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNodeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodeListLogic {
	return &GetNodeListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNodeListLogic) GetNodeList(req *types.GetNodeServerListRequest) (resp *types.GetNodeServerListResponse, err error) {
	tags := make([]string, 0)
	if req.Tags != "" {
		tags = strings.Split(req.Tags, ",")
	}
	total, list, err := l.svcCtx.ServerModel.FindServerListByFilter(l.ctx, &server.ServerFilter{
		Page:   req.Page,
		Size:   req.Size,
		Search: req.Search,
		Tags:   tags,
		Group:  req.GroupId,
	})
	if err != nil {
		l.Errorw("[GetNodeList] Query Database Error: ", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), err.Error())
	}
	nodes := make([]types.Server, 0)
	for _, v := range list {
		node := types.Server{}
		tool.DeepCopy(&node, v)
		// default relay mode
		if node.RelayMode == "" {
			node.RelayMode = "none"
		}
		if len(v.Tags) > 0 {
			if strings.Contains(v.Tags, ",") {
				node.Tags = strings.Split(v.Tags, ",")
			} else {
				node.Tags = []string{v.Tags}
			}
		}
		// parse config
		var cfg map[string]interface{}
		err = json.Unmarshal([]byte(v.Config), &cfg)
		if err != nil {
			cfg = make(map[string]interface{})
		}
		node.Config = cfg
		relayNode := make([]types.NodeRelay, 0)
		err = json.Unmarshal([]byte(v.RelayNode), &relayNode)
		if err != nil {
			l.Errorw("[GetNodeList] Unmarshal RelayNode Error: ", logger.Field("error", err.Error()), logger.Field("relayNode", v.RelayNode))
		}
		node.RelayNode = relayNode
		var status types.NodeStatus
		nodeStatus, err := l.svcCtx.NodeCache.GetNodeStatus(l.ctx, v.Id)
		if err != nil {
			// redis nil is not a Error
			if !errors.Is(err, redis.Nil) {
				l.Errorw("[GetNodeList] Get Node Status Error: ", logger.Field("error", err.Error()))
			}
		} else {
			onlineUser, err := l.svcCtx.NodeCache.GetNodeOnlineUser(l.ctx, v.Id)
			if err != nil {
				l.Errorw("[GetNodeList] Get Node Online User Error: ", logger.Field("error", err.Error()))
			} else {
				status.Online = onlineUser
			}
			status.Cpu = nodeStatus.Cpu
			status.Mem = nodeStatus.Mem
			status.Disk = nodeStatus.Disk
			status.UpdatedAt = nodeStatus.UpdatedAt
		}
		node.Status = &status
		nodes = append(nodes, node)
	}
	return &types.GetNodeServerListResponse{
		Total: total,
		List:  nodes,
	}, nil
}
