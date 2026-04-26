package device

// V4.3:用户删除自己的「加购设备」(is_addon=true)。
//
// 业务规则:
//   - 仅 is_addon=true 的设备可删,套餐基础设备不可删(返回 InvalidParams)
//   - 删除后 user_subscribe.device_count 同步 -1
//   - 不退款(V4.3 决策)
//   - 立即生效:删 server user list 缓存 → 节点下次 pull 失效

import (
	"context"
	"fmt"

	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	subscribelogic "github.com/perfect-panel/server/internal/logic/public/subscribe"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type DeleteAddonDeviceLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAddonDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAddonDeviceLogic {
	return &DeleteAddonDeviceLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DeleteAddonDeviceLogic) DeleteAddonDevice(req *types.DeviceIdRequest) (*types.DeviceStatusResponse, error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	device, err := l.svcCtx.UserModel.FindOneSubscribeDevice(l.ctx, req.DeviceId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find device: %v", err)
	}
	if device.UserId != u.Id {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "device not owned by user")
	}
	if !device.IsAddon {
		return nil, errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams,
				"该设备是套餐基础设备,不允许删除。如需减少设备只能删除「加购设备」"),
			"base device not deletable")
	}

	userSub, err := l.svcCtx.UserModel.FindOneSubscribe(l.ctx, device.UserSubscribeId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find sub: %v", err)
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if e := l.svcCtx.UserModel.DeleteSubscribeDevice(l.ctx, device, tx); e != nil {
			return e
		}
		if userSub.DeviceCount > 0 {
			userSub.DeviceCount--
			if e := l.svcCtx.UserModel.UpdateSubscribe(l.ctx, userSub, tx); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		l.Errorw("[DeleteAddonDevice] tx failed", logger.Field("error", err.Error()))
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "delete addon device: %v", err)
	}

	// 失活节点缓存
	_ = l.svcCtx.Redis.Del(l.ctx, subscribelogic.RedisServerUserListGlobalKey).Err()

	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		UserId:  u.Id,
		Actor:   auditmodel.ActorUser,
		ActorId: u.Id,
		Action:  "delete_addon_device",
		Target:  fmt.Sprintf("device:%d", device.Id),
	}, map[string]interface{}{
		"device_name": device.DeviceName,
	})

	return &types.DeviceStatusResponse{DeviceId: device.Id, Status: 0}, nil
}
