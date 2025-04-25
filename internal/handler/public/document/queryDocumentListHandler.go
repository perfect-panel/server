package document

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/document"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get document list
func QueryDocumentListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := document.NewQueryDocumentListLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryDocumentList()
		result.HttpResult(c, resp, err)
	}
}
