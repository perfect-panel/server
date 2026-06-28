package coupon

import (
	"github.com/perfect-panel/server/internal/logic/admin/coupon"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Delete coupon
func DeleteCouponHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.DeleteCouponRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := coupon.NewDeleteCouponLogic(c.Request.Context(), svcCtx)
		err := l.DeleteCoupon(&req)
		result.HttpResult(c, nil, err)
	}
}
