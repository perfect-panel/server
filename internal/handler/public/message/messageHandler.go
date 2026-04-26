package message

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/message"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

func QueryMessagesHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		l := message.NewQueryMessagesLogic(c, svcCtx)
		resp, err := l.Query()
		result.HttpResult(c, resp, err)
	}
}

func UnreadCountHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		l := message.NewUnreadCountLogic(c, svcCtx)
		resp, err := l.Count()
		result.HttpResult(c, resp, err)
	}
}

func MarkReadHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		l := message.NewMarkReadLogic(c, svcCtx)
		resp, err := l.Mark(id)
		result.HttpResult(c, resp, err)
	}
}

func MarkAllReadHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		l := message.NewMarkAllReadLogic(c, svcCtx)
		resp, err := l.MarkAll()
		result.HttpResult(c, resp, err)
	}
}
