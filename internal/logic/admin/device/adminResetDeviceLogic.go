package device

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	pkgtool "github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

// AdminResetDeviceLogic — force token+uuid rotation. Admin ops bypass user
// hour/day rate limits (operator scenario).
type AdminResetDeviceLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewAdminResetDeviceLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *AdminResetDeviceLogic {
	return &AdminResetDeviceLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminResetDeviceLogic) AdminResetDevice(req *types.AdminDeviceIdRequest) (*types.AdminDeviceStatusResponse, error) {
	d, err := l.svcCtx.UserModel.FindOneSubscribeDevice(l.ctx, req.DeviceId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find device: %v", err)
	}
	now := time.Now()
	d.Token = pkgtool.GenerateDeviceToken()
	d.UUID = pkgtool.GenerateUUIDv4()
	d.LastResetAt = &now
	// Skip incrementing reset_count_hour/day: admin ops shouldn't consume user quota.
	if err := l.svcCtx.UserModel.UpdateSubscribeDevice(l.ctx, d); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update: %v", err)
	}
	svc.InvalidateNodeCache(l.ctx, l.svcCtx)
	adminId, _ := svc.AdminIdFromCtx(l.ctx)
	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		UserId: d.UserId, Actor: auditmodel.ActorAdmin, ActorId: adminId,
		Action: auditmodel.ActionResetDevice,
		Target: fmt.Sprintf("device:%d", d.Id),
	}, map[string]interface{}{"by_admin": true})
	return &types.AdminDeviceStatusResponse{DeviceId: d.Id, Status: d.Status}, nil
}
