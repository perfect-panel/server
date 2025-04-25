package node

import (
	"context"
	"strconv"
	"strings"

	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type GetNodeListLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get Node list
func NewGetNodeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNodeListLogic {
	return &GetNodeListLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNodeListLogic) GetNodeList(req *types.AppUserSubscbribeNodeRequest) (resp *types.AppUserSubscbribeNodeResponse, err error) {
	resp = &types.AppUserSubscbribeNodeResponse{List: make([]types.AppUserSubscbribeNode, 0)}
	userInfo := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	userSubscribe, err := l.svcCtx.UserModel.FindOneUserSubscribe(l.ctx, req.Id)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find user subscribe: %v", err.Error())
	}

	if userInfo.Id != userSubscribe.UserId {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "find user subscribe: %v", err.Error())
	}

	//拿到所有订阅下的服务组id
	var ids []int64
	for _, idStr := range strings.Split(userSubscribe.Subscribe.ServerGroup, ",") {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	//根据服务组id拿到所有节点
	servers, err := l.svcCtx.ServerModel.FindServerListByGroupIds(l.ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, server := range servers {
		resp.List = append(resp.List, types.AppUserSubscbribeNode{
			Id:         server.Id,
			Uuid:       userSubscribe.UUID,
			Traffic:    userSubscribe.Traffic,
			Upload:     userSubscribe.Upload,
			Download:   userSubscribe.Download,
			RelayNode:  server.RelayNode,
			RelayMode:  server.RelayMode,
			Longitude:  server.Longitude,
			Latitude:   server.Latitude,
			Tags:       strings.Split(server.Tags, ","),
			Config:     server.Config,
			ServerAddr: server.ServerAddr,
			Protocol:   server.Protocol,
			SpeedLimit: server.SpeedLimit,
			City:       server.City,
			Country:    server.Country,
			Name:       server.Name,
		})
	}
	return
}
