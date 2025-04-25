package subscribe

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryUserAvailableUserSubscribeLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get Available subscriptions for users
func NewQueryUserAvailableUserSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUserAvailableUserSubscribeLogic {
	return &QueryUserAvailableUserSubscribeLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryUserAvailableUserSubscribeLogic) QueryUserAvailableUserSubscribe(req *types.AppUserSubscribeRequest) (resp *types.AppUserSubscbribeResponse, err error) {
	resp = &types.AppUserSubscbribeResponse{List: make([]types.AppUserSubcbribe, 0)}
	userInfo := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	//查询用户订阅
	subscribeDetails, err := l.svcCtx.UserModel.QueryUserSubscribe(l.ctx, userInfo.Id, 1, 2)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "get query user subscribe error: %v", err.Error())
	}

	userSubscribeMap := make(map[int64]types.AppUserSubcbribe)
	for _, sd := range subscribeDetails {
		userSubscribeInfo := types.AppUserSubcbribe{
			Id:          sd.Id,
			Name:        sd.Subscribe.Name,
			Traffic:     sd.Traffic,
			Upload:      sd.Upload,
			Download:    sd.Download,
			ExpireTime:  sd.ExpireTime.Format(time.DateTime),
			StartTime:   sd.StartTime.Format(time.DateTime),
			DeviceLimit: sd.Subscribe.DeviceLimit,
		}

		//不需要查询节点
		if req.ContainsNodes == nil || !*req.ContainsNodes {
			resp.List = append(resp.List, userSubscribeInfo)
			continue
		}

		//拿到所有订阅下的服务组id
		var ids []int64
		for _, idStr := range strings.Split(sd.Subscribe.ServerGroup, ",") {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				continue
			}
			ids = append(ids, id)
		}
		//根据服务组id拿到所有节点
		servers, err := l.svcCtx.ServerModel.FindServerListByGroupIds(l.ctx, ids)
		if err != nil {
			l.Logger.Errorf("FindServerListByGroupIds error: %v", err.Error())
			continue
		}

		for _, server := range servers {
			userSubscribeInfo.List = append(userSubscribeInfo.List, types.AppUserSubscbribeNode{
				Id:         server.Id,
				Uuid:       sd.UUID,
				Traffic:    sd.Traffic,
				Upload:     sd.Upload,
				Download:   sd.Download,
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
		resp.List = append(resp.List, userSubscribeInfo)
		userSubscribeMap[userSubscribeInfo.Id] = userSubscribeInfo
	}

	for _, userSubscribeInfo := range userSubscribeMap {
		resp.List = append(resp.List, userSubscribeInfo)
	}
	return resp, nil

}
