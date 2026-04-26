package subscribe

// V4.3 我的订阅 — 共享流量池 + N 设备槽视图(决策 14 / 20 / 31)。
// 输出字段对齐前端 6.1 设计稿。

import (
	"context"
	"fmt"
	"time"

	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"

	"github.com/pkg/errors"
)

type QueryMySubscribesLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMySubscribesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMySubscribesLogic {
	return &QueryMySubscribesLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMySubscribesLogic) QueryMySubscribes() (*types.QueryMySubscribesResponse, error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}

	subs, err := l.svcCtx.UserModel.QueryUserSubscribe(l.ctx, u.Id, 1)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query user subscribes: %v", err)
	}

	resp := &types.QueryMySubscribesResponse{List: make([]types.MySubscribeInfo, 0, len(subs))}
	now := time.Now()
	for _, s := range subs {
		// 设备槽位
		devices, _ := l.svcCtx.UserModel.QuerySubscribeDevices(l.ctx, s.Id)
		devList := make([]types.MySubscribeDevice, 0, len(devices))
		for _, d := range devices {
			urls := BuildSubscribeURLs(l.svcCtx, d.Token)
			primary := ""
			if len(urls) > 0 {
				primary = urls[0]
			} else {
				primary = BuildSubscribeURL(l.svcCtx, d.Token)
			}
			devList = append(devList, types.MySubscribeDevice{
				Id:            d.Id,
				DeviceName:    d.DeviceName,
				Token:         d.Token,
				UUID:          d.UUID,
				SubscribeUrl:  primary,
				SubscribeUrls: urls,
				QrCodeUrl:     fmt.Sprintf("/v1/public/qr?token=%s", d.Token),
				LastSeenAt:    timeToMs(d.LastSeenAt),
				LastSeenIp:    d.LastSeenIP,
				TodayTraffic:  d.TodayTraffic,
				Status:        d.Status,
				IsAddon:       d.IsAddon,
			})
		}

		// 用量
		quota := s.Traffic + s.TrafficAddon
		used := s.Download + s.Upload
		var remainPct int64
		if quota > 0 {
			remainPct = (quota - used) * 100 / quota
			if remainPct < 0 {
				remainPct = 0
			}
		}

		// 状态机:正常 / throttled / cutoff
		status := "normal"
		switch {
		case s.CutOffAt != nil && now.After(*s.CutOffAt):
			status = "cutoff"
		case s.ThrottledAt != nil:
			status = "throttled"
		}

		var subName string
		var unitPricePerDevice, maxDeviceCount, addonUnitPrice, addonUnitSize int64
		if s.Subscribe != nil {
			subName = s.Subscribe.Name
			unitPricePerDevice = s.Subscribe.UnitPricePerDevice
			maxDeviceCount = s.Subscribe.MaxDeviceCount
			addonUnitPrice = s.Subscribe.TrafficAddonUnitPrice
			addonUnitSize = s.Subscribe.TrafficAddonUnitSize
		}

		resp.List = append(resp.List, types.MySubscribeInfo{
			Id:                    s.Id,
			SubscribeId:           s.SubscribeId,
			SubscribeName:         subName,
			DeviceCount:           s.DeviceCount,
			TrafficTotal:          quota,
			TrafficAddon:          s.TrafficAddon,
			TrafficUsed:           used,
			TrafficRemainPct:      remainPct,
			StartTime:             s.StartTime.UnixMilli(),
			ExpireTime:            s.ExpireTime.UnixMilli(),
			Status:                status,
			ThrottledAt:           timeToMs(s.ThrottledAt),
			CutOffAt:              timeToMs(s.CutOffAt),
			Devices:               devList,
			UnitPricePerDevice:    unitPricePerDevice,
			MaxDeviceCount:        maxDeviceCount,
			TrafficAddonUnitPrice: addonUnitPrice,
			TrafficAddonUnitSize:  addonUnitSize,
		})
	}
	return resp, nil
}

func timeToMs(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.UnixMilli()
}
