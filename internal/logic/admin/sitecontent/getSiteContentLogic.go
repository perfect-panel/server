package sitecontent

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/model/sitecontent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type GetSiteContentLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewGetSiteContentLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *GetSiteContentLogic {
	return &GetSiteContentLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *GetSiteContentLogic) GetSiteContent(req *types.GetSiteContentRequest) (*types.GetSiteContentResponse, error) {
	lang := strings.TrimSpace(req.Lang)
	var rows []*sitecontent.SiteContent
	var err error
	if req.Prefix != "" {
		rows, err = l.svcCtx.SiteContentModel.ListByPrefix(l.ctx, req.Prefix, lang)
	} else {
		rows, err = l.svcCtx.SiteContentModel.List(l.ctx, lang)
	}
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "list site_content: %v", err)
	}
	resp := &types.GetSiteContentResponse{List: make([]types.SiteContentItem, 0, len(rows))}
	for _, r := range rows {
		resp.List = append(resp.List, types.SiteContentItem{
			Id:          r.Id,
			ContentKey:  r.ContentKey,
			ContentLang: r.ContentLang,
			Title:       r.Title,
			Body:        r.Body,
			Version:     r.Version,
			UpdatedAt:   r.UpdatedAt.UnixMilli(),
		})
	}
	return resp, nil
}
