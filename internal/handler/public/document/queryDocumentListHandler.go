package document

import (
	"github.com/perfect-panel/server/internal/logic/public/document"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get document list
func QueryDocumentListHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := document.NewQueryDocumentListLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryDocumentList()
		result.HttpResult(c, resp, err)
	}
}
