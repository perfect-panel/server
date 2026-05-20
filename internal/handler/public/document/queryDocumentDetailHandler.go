package document

import (
	"github.com/perfect-panel/server/internal/logic/public/document"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get document detail
func QueryDocumentDetailHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.QueryDocumentDetailRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := document.NewQueryDocumentDetailLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryDocumentDetail(&req)
		result.HttpResult(c, resp, err)
	}
}
