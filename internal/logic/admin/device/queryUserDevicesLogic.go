package device

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryUserDevicesLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewQueryUserDevicesLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *QueryUserDevicesLogic {
	return &QueryUserDevicesLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *QueryUserDevicesLogic) QueryUserDevices(req *types.QueryUserDevicesRequest) (*types.QueryUserDevicesResponse, error) {
	if req.UserId == 0 {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "user_id required")
	}
	var devices []*user.SubscribeDevice
	conn := l.svcCtx.DB.WithContext(l.ctx).Model(&user.SubscribeDevice{}).Where("user_id = ?", req.UserId)
	if req.UserSubscribeId > 0 {
		conn = conn.Where("user_subscribe_id = ?", req.UserSubscribeId)
	}
	if err := conn.Order("id ASC").Find(&devices).Error; err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query devices: %v", err)
	}
	resp := &types.QueryUserDevicesResponse{List: make([]types.AdminUserDeviceItem, 0, len(devices))}
	for _, d := range devices {
		var lastSeen int64
		if d.LastSeenAt != nil {
			lastSeen = d.LastSeenAt.UnixMilli()
		}
		resp.List = append(resp.List, types.AdminUserDeviceItem{
			Id:              d.Id,
			UserSubscribeId: d.UserSubscribeId,
			UserId:          d.UserId,
			DeviceName:      d.DeviceName,
			Token:           d.Token,
			UUID:            d.UUID,
			LastSeenIp:      d.LastSeenIP,
			LastSeenAt:      lastSeen,
			TodayTraffic:    d.TodayTraffic,
			Status:          d.Status,
			IsAddon:         d.IsAddon,
		})
	}
	return resp, nil
}
