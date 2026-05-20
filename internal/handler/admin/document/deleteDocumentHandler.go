package document

import (
	"github.com/perfect-panel/server/internal/logic/admin/document"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Delete document
func DeleteDocumentHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.DeleteDocumentRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := document.NewDeleteDocumentLogic(c.Request.Context(), svcCtx)
		err := l.DeleteDocument(&req)
		result.HttpResult(c, nil, err)
	}
}
