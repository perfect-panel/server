package common

import (
	"context"
	"strings"

	"github.com/perfect-panel/server/internal/model/sitecontent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

// GetSiteContentItemLogic — V4.3 决策 25 公开接口:
// 用户端按 (key, lang) 拉单条 site_content 内容(客户端使用教程 / 用户协议 等)。
// 缺失目标语言时回退 zh-CN。
type GetSiteContentItemLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSiteContentItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSiteContentItemLogic {
	return &GetSiteContentItemLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSiteContentItemLogic) GetSiteContentItem(req *types.GetSiteContentItemRequest) (*types.GetSiteContentItemResponse, error) {
	key := strings.TrimSpace(req.Key)
	if key == "" {
		return nil, errors.Wrap(xerr.NewErrCode(xerr.InvalidParams), "key is required")
	}
	lang := strings.TrimSpace(req.Lang)
	if lang == "" {
		lang = sitecontent.DefaultLang
	}
	row, err := l.svcCtx.SiteContentModel.GetWithFallback(l.ctx, key, lang)
	if err != nil {
		// 找不到也不报错,返回空内容(前端弹层自然显示"暂无教程")
		return &types.GetSiteContentItemResponse{
			ContentKey:  key,
			ContentLang: lang,
		}, nil
	}
	return &types.GetSiteContentItemResponse{
		ContentKey:  row.ContentKey,
		ContentLang: row.ContentLang,
		Title:       row.Title,
		Body:        row.Body,
		Version:     row.Version,
		UpdatedAt:   row.UpdatedAt.UnixMilli(),
	}, nil
}
