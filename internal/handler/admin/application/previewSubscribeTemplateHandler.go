package application

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/application"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Preview Template
func PreviewSubscribeTemplateHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.PreviewSubscribeTemplateRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := application.NewPreviewSubscribeTemplateLogic(c.Request.Context(), svcCtx)
		resp, err := l.PreviewSubscribeTemplate(&req)
		result.HttpResult(c, resp, err)

	}
}
