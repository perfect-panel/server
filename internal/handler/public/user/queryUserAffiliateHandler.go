package user

import (
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Query User Affiliate Count
func QueryUserAffiliateHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := user.NewQueryUserAffiliateLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryUserAffiliate()
		result.HttpResult(c, resp, err)
	}
}
