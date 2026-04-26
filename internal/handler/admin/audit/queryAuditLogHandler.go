package audit

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/audit"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// QueryAuditLogHandler Query audit log.
func QueryAuditLogHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.QueryAuditLogRequest
		_ = c.ShouldBindQuery(&req)
		l := audit.NewQueryAuditLogLogic(c, svcCtx)
		resp, err := l.QueryAuditLog(&req)
		result.HttpResult(c, resp, err)
	}
}
