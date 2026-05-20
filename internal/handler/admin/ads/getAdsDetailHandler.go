package ads

import (
	"github.com/perfect-panel/server/internal/logic/admin/ads"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get Ads Detail
func GetAdsDetailHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.GetAdsDetailRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		ctx := c.Request.Context()
		l := ads.NewGetAdsDetailLogic(ctx, svcCtx)
		resp, err := l.GetAdsDetail(&req)
		result.HttpResult(c, resp, err)
	}
}
