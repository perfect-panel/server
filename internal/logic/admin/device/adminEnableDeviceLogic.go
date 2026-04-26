package device

import (
	"fmt"

	"github.com/gin-gonic/gin"
	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	pkgtool "github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

// AdminEnableDeviceLogic — enable + re-issue token+uuid (matches user-side Enable, V4.3 decision 16).
type AdminEnableDeviceLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewAdminEnableDeviceLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *AdminEnableDeviceLogic {
	return &AdminEnableDeviceLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminEnableDeviceLogic) AdminEnableDevice(req *types.AdminDeviceIdRequest) (*types.AdminDeviceStatusResponse, error) {
	d, err := l.svcCtx.UserModel.FindOneSubscribeDevice(l.ctx, req.DeviceId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find device: %v", err)
	}
	d.Status = 1
	d.Token = pkgtool.GenerateDeviceToken()
	d.UUID = pkgtool.GenerateUUIDv4()
	if err := l.svcCtx.UserModel.UpdateSubscribeDevice(l.ctx, d); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update: %v", err)
	}
	svc.InvalidateNodeCache(l.ctx, l.svcCtx)
	adminId, _ := svc.AdminIdFromCtx(l.ctx)
	_ = l.svcCtx.AuditModel.Append(l.ctx, &auditmodel.AuditLog{
		UserId: d.UserId, Actor: auditmodel.ActorAdmin, ActorId: adminId,
		Action: auditmodel.ActionEnableDevice,
		Target: fmt.Sprintf("device:%d", d.Id),
	})
	return &types.AdminDeviceStatusResponse{DeviceId: d.Id, Status: d.Status}, nil
}
