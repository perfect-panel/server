package sitecontent

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/sitecontent"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// UpsertSiteContentHandler Upsert a site content row.
func UpsertSiteContentHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.UpsertSiteContentRequest
		_ = c.ShouldBind(&req)
		l := sitecontent.NewUpsertSiteContentLogic(c, svcCtx)
		resp, err := l.UpsertSiteContent(&req)
		result.HttpResult(c, resp, err)
	}
}
