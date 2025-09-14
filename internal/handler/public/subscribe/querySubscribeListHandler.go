package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// QuerySubscribeListHandler Get subscribe list
func QuerySubscribeListHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.QuerySubscribeListRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := subscribe.NewQuerySubscribeListLogic(c.Request.Context(), svcCtx)
		resp, err := l.QuerySubscribeList(&req)
		result.HttpResult(c, resp, err)
	}
}
