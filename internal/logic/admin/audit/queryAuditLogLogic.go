package audit

import (
	"time"

	"github.com/gin-gonic/gin"
	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type QueryAuditLogLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewQueryAuditLogLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *QueryAuditLogLogic {
	return &QueryAuditLogLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *QueryAuditLogLogic) QueryAuditLog(req *types.QueryAuditLogRequest) (*types.QueryAuditLogResponse, error) {
	filter := &auditmodel.Filter{Action: req.Action, Actor: req.Actor}
	if req.UserId != 0 {
		v := req.UserId
		filter.UserId = &v
	}
	if req.ActorId != 0 {
		v := req.ActorId
		filter.ActorId = &v
	}
	if req.Since != 0 {
		t := time.UnixMilli(req.Since)
		filter.Since = &t
	}
	if req.Until != 0 {
		t := time.UnixMilli(req.Until)
		filter.Until = &t
	}
	rows, total, err := l.svcCtx.AuditModel.Query(l.ctx, filter, req.Page, req.Size)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query audit: %v", err)
	}
	resp := &types.QueryAuditLogResponse{Total: total, List: make([]types.AuditLogItem, 0, len(rows))}
	for _, r := range rows {
		resp.List = append(resp.List, types.AuditLogItem{
			Id:        r.Id,
			UserId:    r.UserId,
			Actor:     r.Actor,
			ActorId:   r.ActorId,
			Action:    r.Action,
			Target:    r.Target,
			Detail:    r.Detail,
			ClientIp:  r.ClientIp,
			CreatedAt: r.CreatedAt.UnixMilli(),
		})
	}
	return resp, nil
}
