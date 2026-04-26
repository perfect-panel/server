package message

// V4.3 站内信用户侧:列表 / 未读计数 / 标记已读 / 全部已读。

import (
	"strconv"

	"github.com/gin-gonic/gin"
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

// QueryMessages — GET /v1/portal/messages?page=&size=&unread=1
type QueryMessagesLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMessagesLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *QueryMessagesLogic {
	return &QueryMessagesLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

type queryMessagesResp struct {
	List  []messageItem `json:"list"`
	Total int64         `json:"total"`
}
type messageItem struct {
	Id        int64  `json:"id"`
	Category  string `json:"category"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Link      string `json:"link,omitempty"`
	ReadAt    int64  `json:"read_at"` // 0 = unread
	CreatedAt int64  `json:"created_at"`
}

func (l *QueryMessagesLogic) Query() (*queryMessagesResp, error) {
	u, ok := currentUser(l.ctx)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	page, _ := strconv.Atoi(l.ctx.Query("page"))
	size, _ := strconv.Atoi(l.ctx.Query("size"))
	unread := l.ctx.Query("unread") == "1"
	rows, total, err := l.svcCtx.MessageModel.List(l.ctx.Request.Context(), u.Id, page, size, unread)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "list messages: %v", err)
	}
	resp := &queryMessagesResp{Total: total, List: make([]messageItem, 0, len(rows))}
	for _, r := range rows {
		var readMs int64
		if r.ReadAt != nil {
			readMs = r.ReadAt.UnixMilli()
		}
		resp.List = append(resp.List, messageItem{
			Id:        r.Id,
			Category:  r.Category,
			Title:     r.Title,
			Body:      r.Body,
			Link:      r.Link,
			ReadAt:    readMs,
			CreatedAt: r.CreatedAt.UnixMilli(),
		})
	}
	return resp, nil
}

// UnreadCount — GET /v1/portal/messages/unread_count
type UnreadCountLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewUnreadCountLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *UnreadCountLogic {
	return &UnreadCountLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *UnreadCountLogic) Count() (map[string]int64, error) {
	u, ok := currentUser(l.ctx)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	n, err := l.svcCtx.MessageModel.UnreadCount(l.ctx.Request.Context(), u.Id)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "unread count: %v", err)
	}
	return map[string]int64{"unread": n}, nil
}

// MarkRead — PUT /v1/portal/messages/:id/read
type MarkReadLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewMarkReadLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *MarkReadLogic) Mark(id int64) (map[string]int64, error) {
	u, ok := currentUser(l.ctx)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	if err := l.svcCtx.MessageModel.MarkRead(l.ctx.Request.Context(), u.Id, id); err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "mark read: %v", err)
	}
	return map[string]int64{"id": id}, nil
}

// MarkAllRead — POST /v1/portal/messages/read_all
type MarkAllReadLogic struct {
	logger.Logger
	ctx    *gin.Context
	svcCtx *svc.ServiceContext
}

func NewMarkAllReadLogic(ctx *gin.Context, svcCtx *svc.ServiceContext) *MarkAllReadLogic {
	return &MarkAllReadLogic{Logger: logger.WithContext(ctx.Request.Context()), ctx: ctx, svcCtx: svcCtx}
}

func (l *MarkAllReadLogic) MarkAll() (map[string]int64, error) {
	u, ok := currentUser(l.ctx)
	if !ok {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "user context missing")
	}
	n, err := l.svcCtx.MessageModel.MarkAllRead(l.ctx.Request.Context(), u.Id)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.ERROR), "mark all read: %v", err)
	}
	return map[string]int64{"updated": n}, nil
}

