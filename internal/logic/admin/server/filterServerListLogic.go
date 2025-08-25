package server

import (
	"context"

	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type FilterServerListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFilterServerListLogic Filter Server List
func NewFilterServerListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilterServerListLogic {
	return &FilterServerListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FilterServerListLogic) FilterServerList(req *types.FilterServerListRequest) (resp *types.FilterServerListResponse, err error) {
	total, data, err := l.svcCtx.NodeModel.FilterServerList(l.ctx, &node.FilterParams{
		Page:   req.Page,
		Size:   req.Size,
		Search: req.Search,
	})
	if err != nil {
		l.Errorw("[FilterServerList] Query Database Error: ", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "[FilterServerList] Query Database Error")
	}

	list := make([]types.Server, 0)

	for _, datum := range data {
		var server types.Server
		tool.DeepCopy(&server, datum)

		// handler protocols
		var protocols []types.Protocol
		dst, err := datum.UnmarshalProtocols()
		if err != nil {
			l.Errorf("[FilterServerList] UnmarshalProtocols Error: %s", err.Error())
			continue
		}
		tool.DeepCopy(&protocols, dst)
		server.Protocols = protocols
		// handler status
		server.Status = l.handlerServerStatus(datum.Id)
		list = append(list, server)
	}

	return &types.FilterServerListResponse{
		List:  list,
		Total: total,
	}, nil
}

func (l *FilterServerListLogic) handlerServerStatus(id int64) types.ServerStatus {
	var result types.ServerStatus
	nodeStatus, err := l.svcCtx.NodeCache.GetNodeStatus(l.ctx, id)
	if err != nil {
		l.Errorw("[handlerServerStatus] GetNodeStatus Error: ", logger.Field("error", err.Error()), logger.Field("node_id", id))
		return result
	}
	result = types.ServerStatus{
		Mem:    nodeStatus.Mem,
		Cpu:    nodeStatus.Cpu,
		Disk:   nodeStatus.Disk,
		Online: make([]types.ServerOnlineUser, 0),
	}

	// parse online users
	onlineUser, err := l.svcCtx.NodeCache.GetNodeOnlineUser(l.ctx, id)
	if err != nil {
		l.Errorw("[handlerServerStatus] GetNodeOnlineUser Error: ", logger.Field("error", err.Error()), logger.Field("node_id", id))
		return result
	}

	var onlineList []types.ServerOnlineUser
	var onlineMap = make(map[int64]types.ServerOnlineUser)
	// group by user_id
	for subId, info := range onlineUser {
		data, err := l.svcCtx.UserModel.FindOneUserSubscribe(l.ctx, subId)
		if err != nil {
			l.Errorw("[handlerServerStatus] FindOneSubscribe Error: ", logger.Field("error", err.Error()))
			continue
		}
		if online, exist := onlineMap[data.UserId]; !exist {
			onlineMap[data.UserId] = types.ServerOnlineUser{
				IP:          info,
				UserId:      data.UserId,
				Subscribe:   data.Subscribe.Name,
				SubscribeId: data.SubscribeId,
				Traffic:     data.Traffic,
				ExpiredAt:   data.ExpireTime.UnixMilli(),
			}
		} else {
			online.IP = append(online.IP, info...)
			onlineMap[data.UserId] = online
		}
	}

	for _, online := range onlineMap {
		onlineList = append(onlineList, online)
	}

	result.Online = onlineList

	return result
}
