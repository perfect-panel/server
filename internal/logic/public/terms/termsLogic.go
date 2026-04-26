package terms

// V4.4 #45 用户协议版本管理。
// - GET  /v1/portal/terms/status?lang=zh-CN  → { current_version, accepted_version, needs_accept, title, body }
// - POST /v1/portal/terms/accept              → 把 user.terms_version 写为当前 version

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/model/sitecontent"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"

	"github.com/pkg/errors"
)

func currentUser(c *gin.Context) (*user.User, bool) {
	u, ok := c.Request.Context().Value(constant.CtxKeyUser).(*user.User)
	return u, ok
}

type StatusResponse struct {
	CurrentVersion  string `json:"current_version"`
	AcceptedVersion string `json:"accepted_version"`
	NeedsAccept     bool   `json:"needs_accept"`
	Title           string `json:"title"`
	Body            string `json:"body"`
	Lang            string `json:"lang"`
}

// StatusLogic — 即使用户尚未接受,也允许查询(展示协议内容)。
type StatusLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewStatusLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *StatusLogic {
	return &StatusLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *StatusLogic) Status() (*StatusResponse, error) {
	u, ok := currentUser(l.ctx)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	lang := l.ctx.Query("lang")
	if lang == "" {
		lang = sitecontent.DefaultLang
	}
	row, err := l.svcCtx.SiteContentModel.GetWithFallback(l.ctx.Request.Context(), sitecontent.KeyTermsOfUse, lang)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "load terms: %v", err)
	}
	resp := &StatusResponse{
		CurrentVersion:  row.Version,
		AcceptedVersion: u.TermsVersion,
		Title:           row.Title,
		Body:            row.Body,
		Lang:            row.ContentLang,
	}
	resp.NeedsAccept = resp.CurrentVersion != "" && resp.CurrentVersion != resp.AcceptedVersion
	return resp, nil
}

type AcceptResponse struct {
	AcceptedVersion string `json:"accepted_version"`
}

// AcceptLogic — 把当前 lang 的 version 写到 user.terms_version。
type AcceptLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewAcceptLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *AcceptLogic {
	return &AcceptLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *AcceptLogic) Accept() (*AcceptResponse, error) {
	u, ok := currentUser(l.ctx)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	lang := l.ctx.Query("lang")
	if lang == "" {
		lang = sitecontent.DefaultLang
	}
	row, err := l.svcCtx.SiteContentModel.GetWithFallback(l.ctx.Request.Context(), sitecontent.KeyTermsOfUse, lang)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "load terms: %v", err)
	}
	u.TermsVersion = row.Version
	if err := l.svcCtx.UserModel.Update(l.ctx.Request.Context(), u); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "update user: %v", err)
	}
	return &AcceptResponse{AcceptedVersion: row.Version}, nil
}
