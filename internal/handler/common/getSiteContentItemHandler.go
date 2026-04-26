package common

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/common"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// GetSiteContentItemHandler — V4.3 决策 25 公开端点:
// GET /v1/common/site_content?key=client_tutorial_xxx&lang=zh-CN
func GetSiteContentItemHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.GetSiteContentItemRequest
		_ = c.ShouldBindQuery(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}
		l := common.NewGetSiteContentItemLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetSiteContentItem(&req)
		result.HttpResult(c, resp, err)
	}
}
