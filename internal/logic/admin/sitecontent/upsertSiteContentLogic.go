package sitecontent

import (
	"fmt"

	"github.com/gin-gonic/gin"
	auditmodel "github.com/perfect-panel/server/internal/model/audit"
	"github.com/perfect-panel/server/internal/model/sitecontent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type UpsertSiteContentLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewUpsertSiteContentLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *UpsertSiteContentLogic {
	return &UpsertSiteContentLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *UpsertSiteContentLogic) UpsertSiteContent(req *types.UpsertSiteContentRequest) (*types.UpsertSiteContentResponse, error) {
	if req.ContentKey == "" {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidParams), "content_key required")
	}
	if req.ContentLang == "" {
		req.ContentLang = sitecontent.DefaultLang
	}
	row := &sitecontent.SiteContent{
		ContentKey:  req.ContentKey,
		ContentLang: req.ContentLang,
		Title:       req.Title,
		Body:        req.Body,
		Version:     req.Version, // V4.4 #45: empty preserves prior version, non-empty bumps
	}
	if err := l.svcCtx.SiteContentModel.Upsert(l.ctx, row); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "upsert: %v", err)
	}
	adminId, _ := svc.AdminIdFromCtx(l.ctx)
	_ = l.svcCtx.AuditModel.AppendDetail(l.ctx, &auditmodel.AuditLog{
		Actor:   auditmodel.ActorAdmin,
		ActorId: adminId,
		Action:  "cms_upsert",
		Target:  fmt.Sprintf("site_content:%s:%s", req.ContentKey, req.ContentLang),
	}, map[string]interface{}{"title": req.Title, "version": req.Version})
	return &types.UpsertSiteContentResponse{Id: row.Id}, nil
}
