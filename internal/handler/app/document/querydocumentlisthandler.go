package document

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/app/document"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Get document list
func QueryDocumentListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := document.NewQueryDocumentListLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryDocumentList()
		result.HttpResult(c, resp, err)
	}
}
