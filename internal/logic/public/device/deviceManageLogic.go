package device

// V4.3 设备槽管理:重置 / 停用 / 启用 / 改名 / 一键重置全部(决策 9 / 11 / 16 / 26 / 34)。
//
// 频率限制(决策 9):
//   - reset_count_hour < 3
//   - reset_count_day  < 10
//   - last_reset_at 跨越整点/整日时计数器自动归零
//
// 节点同步(决策 26):
//   - 重置后 DEL `server:user:` 全部 key,节点 ≤ pull_interval 自然刷新
//   - SLA:60s 内收敛(节点默认 60s 拉一次)

import (
	"context"
	"fmt"
	"time"

	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	subscribelogic "github.com/perfect-panel/server/internal/logic/public/subscribe"
	"github.com/perfect-panel/server/internal/model/node"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	resetMaxPerHour = 3
	resetMaxPerDay  = 10
	// 一键重置冷却:30 min(决策 4.6 — 防止扫枪)
	resetAllCooldown = 30 * time.Minute
	// 一键重置冷却 Redis key
	resetAllCooldownKey = "subscribe:reset_all:cooldown:%d"
)

// =================== ResetDevice =================== //

type ResetDeviceLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetDeviceLogic {
	return &ResetDeviceLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ResetDeviceLogic) ResetDevice(req *types.DeviceIdRequest) (*types.DeviceResetResponse, error) {
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

	now := time.Now()
	if err := applyResetFrequencyCheck(device, now); err != nil {
		return nil, err
	}

	device.Token = tool.GenerateDeviceToken()
	device.UUID = tool.GenerateUUIDv4()
	device.LastResetAt = &now
	device.ResetCountHour++
	device.ResetCountDay++

	if err := l.svcCtx.UserModel.UpdateSubscribeDevice(l.ctx, device); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update device: %v", err)
	}

	invalidateAllServerUserListCache(l.ctx, l.svcCtx)

	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		UserId:  u.Id,
		Actor:   auditmodel.ActorUser,
		ActorId: u.Id,
		Action:  auditmodel.ActionResetDevice,
		Target:  fmt.Sprintf("device:%d", device.Id),
	}, map[string]interface{}{
		"reset_count_hour": device.ResetCountHour,
		"reset_count_day":  device.ResetCountDay,
	})

	urls := subscribelogic.BuildSubscribeURLs(l.svcCtx, device.Token)
	primary := ""
	if len(urls) > 0 {
		primary = urls[0]
	} else {
		primary = subscribelogic.BuildSubscribeURL(l.svcCtx, device.Token)
	}
	return &types.DeviceResetResponse{
		DeviceId:       device.Id,
		Token:          device.Token,
		UUID:           device.UUID,
		SubscribeUrl:   primary,
		SubscribeUrls:  urls,
		ResetCountHour: device.ResetCountHour,
		ResetCountDay:  device.ResetCountDay,
	}, nil
}

// applyResetFrequencyCheck 检查并按需归零计数器。
// 跨小时 → reset_count_hour=0;跨自然日 → reset_count_day=0(同时归零 hour)。
//
// 触发限制时返回带中文人话提示的 InvalidParams,前端会直接 toast 出来,
// 而不是显示通用的「请求参数不正确」。
func applyResetFrequencyCheck(device *user.SubscribeDevice, now time.Time) error {
	if device.LastResetAt != nil {
		last := *device.LastResetAt
		if !sameDay(last, now) {
			device.ResetCountDay = 0
			device.ResetCountHour = 0
		} else if !sameHour(last, now) {
			device.ResetCountHour = 0
		}
	}
	if device.ResetCountHour >= resetMaxPerHour {
		// 本小时已用满 3 次,显示离下一小时还有多久 + 具体可用时间点
		nextHour := now.Truncate(time.Hour).Add(time.Hour)
		wait := nextHour.Sub(now)
		return errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams,
				fmt.Sprintf("本小时换设备已达上限(%d 次),请于 %s 后重试(%s 后)",
					resetMaxPerHour, nextHour.Format("15:04"), formatDuration(wait))),
			"hourly rate limit")
	}
	if device.ResetCountDay >= resetMaxPerDay {
		// 当天已用满 10 次,显示明天 0 点可用 + 还要等多久
		nextDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		wait := nextDay.Sub(now)
		return errors.Wrap(
			xerr.NewErrCodeMsg(xerr.InvalidParams,
				fmt.Sprintf("今日换设备已达上限(%d 次),请于明日 00:00 后重试(还需 %s)",
					resetMaxPerDay, formatDuration(wait))),
			"daily rate limit")
	}
	return nil
}

// formatDuration — 把 time.Duration 渲染成 "X 小时 Y 分钟" / "Y 分 Z 秒"
// 之类的人话,供错误提示使用。
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0 秒"
	}
	hours := int(d / time.Hour)
	minutes := int((d % time.Hour) / time.Minute)
	seconds := int((d % time.Minute) / time.Second)
	switch {
	case hours > 0:
		return fmt.Sprintf("%d 小时 %d 分钟", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%d 分 %d 秒", minutes, seconds)
	default:
		return fmt.Sprintf("%d 秒", seconds)
	}
}

func sameHour(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay() && a.Hour() == b.Hour()
}
func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

// invalidateAllServerUserListCache: DEL 所有 `server:user:*`,通知所有节点重新拉取 user list。
// 决策 26 SLA:节点默认 pull_interval=60s,最坏 60s 内收敛。
func invalidateAllServerUserListCache(ctx context.Context, svcCtx *svc.ServiceContext) {
	pattern := node.ServerUserListCacheKey + "*"
	iter := svcCtx.Redis.Scan(ctx, 0, pattern, 100).Iterator()
	keys := make([]string, 0, 16)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		logger.WithContext(ctx).Error("[invalidateAllServerUserListCache] scan failed",
			logger.Field("error", err.Error()), logger.Field("pattern", pattern))
		return
	}
	if len(keys) == 0 {
		return
	}
	if err := svcCtx.Redis.Del(ctx, keys...).Err(); err != nil {
		logger.WithContext(ctx).Error("[invalidateAllServerUserListCache] del failed",
			logger.Field("error", err.Error()), logger.Field("count", len(keys)))
	}
}

// =================== DisableDevice =================== //

type DisableDeviceLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDisableDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableDeviceLogic {
	return &DisableDeviceLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DisableDeviceLogic) DisableDevice(req *types.DeviceIdRequest) (*types.DeviceStatusResponse, error) {
	device, u, err := loadOwnedDevice(l.ctx, l.svcCtx, req.DeviceId)
	if err != nil {
		return nil, err
	}
	device.Status = 0
	if err := l.svcCtx.UserModel.UpdateSubscribeDevice(l.ctx, device); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update device: %v", err)
	}
	invalidateAllServerUserListCache(l.ctx, l.svcCtx)
	_ = l.svcCtx.AuditModel.Append(l.ctx, &auditmodel.AuditLog{
		UserId: u.Id, Actor: auditmodel.ActorUser, ActorId: u.Id,
		Action: auditmodel.ActionDisableDevice,
		Target: fmt.Sprintf("device:%d", device.Id),
	})
	return &types.DeviceStatusResponse{DeviceId: device.Id, Status: device.Status}, nil
}

// =================== EnableDevice =================== //

type EnableDeviceLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEnableDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnableDeviceLogic {
	return &EnableDeviceLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// 启用设备时换发新 token + uuid(决策 16:防停用期间 URL 泄露后重启即生效)。
func (l *EnableDeviceLogic) EnableDevice(req *types.DeviceIdRequest) (*types.DeviceStatusResponse, error) {
	device, u, err := loadOwnedDevice(l.ctx, l.svcCtx, req.DeviceId)
	if err != nil {
		return nil, err
	}
	device.Status = 1
	device.Token = tool.GenerateDeviceToken()
	device.UUID = tool.GenerateUUIDv4()
	if err := l.svcCtx.UserModel.UpdateSubscribeDevice(l.ctx, device); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update device: %v", err)
	}
	invalidateAllServerUserListCache(l.ctx, l.svcCtx)
	_ = l.svcCtx.AuditModel.Append(l.ctx, &auditmodel.AuditLog{
		UserId: u.Id, Actor: auditmodel.ActorUser, ActorId: u.Id,
		Action: auditmodel.ActionEnableDevice,
		Target: fmt.Sprintf("device:%d", device.Id),
	})
	return &types.DeviceStatusResponse{DeviceId: device.Id, Status: device.Status}, nil
}

// =================== RenameDevice =================== //

type RenameDeviceLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRenameDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RenameDeviceLogic {
	return &RenameDeviceLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *RenameDeviceLogic) RenameDevice(req *types.DeviceRenameRequest) (*types.DeviceRenameResponse, error) {
	device, u, err := loadOwnedDevice(l.ctx, l.svcCtx, req.DeviceId)
	if err != nil {
		return nil, err
	}
	name := req.Name
	if len(name) == 0 || len(name) > 64 {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "device name must be 1-64 chars")
	}
	device.DeviceName = name
	if err := l.svcCtx.UserModel.UpdateSubscribeDevice(l.ctx, device); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update device: %v", err)
	}
	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		UserId: u.Id, Actor: auditmodel.ActorUser, ActorId: u.Id,
		Action: auditmodel.ActionRenameDevice,
		Target: fmt.Sprintf("device:%d", device.Id),
	}, map[string]interface{}{"name": name})
	return &types.DeviceRenameResponse{DeviceId: device.Id, Name: device.DeviceName}, nil
}

// =================== ResetAllDevices =================== //

type ResetAllDevicesLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetAllDevicesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetAllDevicesLogic {
	return &ResetAllDevicesLogic{Logger: logger.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ResetAllDevicesLogic) ResetAllDevices(req *types.ResetAllDevicesRequest) (*types.ResetAllDevicesResponse, error) {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	userSub, err := l.svcCtx.UserModel.FindOneSubscribe(l.ctx, req.UserSubscribeId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find subscribe: %v", err)
	}
	if userSub.UserId != u.Id {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "subscribe not owned by user")
	}

	// 30min 冷却 — Redis SETNX 原子性最强
	cdKey := fmt.Sprintf(resetAllCooldownKey, userSub.Id)
	ok2, err := l.svcCtx.Redis.SetNX(l.ctx, cdKey, "1", resetAllCooldown).Result()
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "redis cooldown: %v", err)
	}
	if !ok2 {
		ttl, _ := l.svcCtx.Redis.TTL(l.ctx, cdKey).Result()
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams),
			"reset_all in cooldown, retry in %ds", int(ttl.Seconds()))
	}

	devices, err := l.svcCtx.UserModel.QuerySubscribeDevices(l.ctx, userSub.Id)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query devices: %v", err)
	}
	now := time.Now()
	resetCount := 0
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		for _, d := range devices {
			if d.Status != 1 {
				continue
			}
			// 频率检查在这里宽松处理:一键重置整体计 1 次到每个 device(决策保留可控成本)
			if err := applyResetFrequencyCheck(d, now); err != nil {
				continue
			}
			d.Token = tool.GenerateDeviceToken()
			d.UUID = tool.GenerateUUIDv4()
			d.LastResetAt = &now
			d.ResetCountHour++
			d.ResetCountDay++
			if err := l.svcCtx.UserModel.UpdateSubscribeDevice(l.ctx, d, tx); err != nil {
				return err
			}
			resetCount++
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "reset all tx: %v", err)
	}
	invalidateAllServerUserListCache(l.ctx, l.svcCtx)
	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		UserId: u.Id, Actor: auditmodel.ActorUser, ActorId: u.Id,
		Action: auditmodel.ActionResetAllDevices,
		Target: fmt.Sprintf("user_subscribe:%d", userSub.Id),
	}, map[string]interface{}{"reset_count": resetCount})
	return &types.ResetAllDevicesResponse{ResetCount: resetCount}, nil
}

// loadOwnedDevice helper:加载设备并校验归属。
func loadOwnedDevice(ctx context.Context, svcCtx *svc.ServiceContext, deviceId int64) (*user.SubscribeDevice, *user.User, error) {
	u, ok := ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		return nil, nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	device, err := svcCtx.UserModel.FindOneSubscribeDevice(ctx, deviceId)
	if err != nil {
		return nil, nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find device: %v", err)
	}
	if device.UserId != u.Id {
		return nil, nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "device not owned by user")
	}
	return device, u, nil
}
