package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Query User Affiliate Count
func QueryUserAffiliateHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := user.NewQueryUserAffiliateLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryUserAffiliate()
		result.HttpResult(c, resp, err)
	}
}
