package document

import (
	"github.com/perfect-panel/server/internal/logic/admin/document"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Batch delete document
func BatchDeleteDocumentHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.BatchDeleteDocumentRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := document.NewBatchDeleteDocumentLogic(c.Request.Context(), svcCtx)
		err := l.BatchDeleteDocument(&req)
		result.HttpResult(c, nil, err)
	}
}
