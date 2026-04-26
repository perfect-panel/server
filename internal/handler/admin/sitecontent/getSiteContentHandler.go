package sitecontent

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/sitecontent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// GetSiteContentHandler List site content rows (CMS).
func GetSiteContentHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetSiteContentRequest
		_ = c.ShouldBindQuery(&req)
		l := sitecontent.NewGetSiteContentLogic(c, svcCtx)
		resp, err := l.GetSiteContent(&req)
		result.HttpResult(c, resp, err)
	}
}
