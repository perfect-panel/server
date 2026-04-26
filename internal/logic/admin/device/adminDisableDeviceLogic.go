package device

import (
	"fmt"

	"github.com/gin-gonic/gin"
	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type AdminDisableDeviceLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewAdminDisableDeviceLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *AdminDisableDeviceLogic {
	return &AdminDisableDeviceLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminDisableDeviceLogic) AdminDisableDevice(req *types.AdminDeviceIdRequest) (*types.AdminDeviceStatusResponse, error) {
	d, err := l.svcCtx.UserModel.FindOneSubscribeDevice(l.ctx, req.DeviceId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find device: %v", err)
	}
	d.Status = 0
	if err := l.svcCtx.UserModel.UpdateSubscribeDevice(l.ctx, d); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update: %v", err)
	}
	svc.InvalidateNodeCache(l.ctx, l.svcCtx)
	adminId, _ := svc.AdminIdFromCtx(l.ctx)
	_ = l.svcCtx.AuditModel.Append(l.ctx, &auditmodel.AuditLog{
		UserId: d.UserId, Actor: auditmodel.ActorAdmin, ActorId: adminId,
		Action: auditmodel.ActionDisableDevice,
		Target: fmt.Sprintf("device:%d", d.Id),
	})
	return &types.AdminDeviceStatusResponse{DeviceId: d.Id, Status: d.Status}, nil
}
