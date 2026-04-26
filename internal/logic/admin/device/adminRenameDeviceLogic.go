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

// AdminRenameDeviceLogic — rename to a human-readable label such as
// "user A's old iPhone".
type AdminRenameDeviceLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewAdminRenameDeviceLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *AdminRenameDeviceLogic {
	return &AdminRenameDeviceLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *AdminRenameDeviceLogic) AdminRenameDevice(req *types.AdminDeviceRenameRequest) (*types.AdminDeviceRenameResponse, error) {
	d, err := l.svcCtx.UserModel.FindOneSubscribeDevice(l.ctx, req.DeviceId)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "find device: %v", err)
	}
	if len(req.Name) == 0 || len(req.Name) > 64 {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "name must be 1-64 chars")
	}
	d.DeviceName = req.Name
	if err := l.svcCtx.UserModel.UpdateSubscribeDevice(l.ctx, d); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update: %v", err)
	}
	adminId, _ := svc.AdminIdFromCtx(l.ctx)
	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		UserId: d.UserId, Actor: auditmodel.ActorAdmin, ActorId: adminId,
		Action: auditmodel.ActionRenameDevice,
		Target: fmt.Sprintf("device:%d", d.Id),
	}, map[string]interface{}{"name": req.Name, "by_admin": true})
	return &types.AdminDeviceRenameResponse{DeviceId: d.Id, Name: d.DeviceName}, nil
}
