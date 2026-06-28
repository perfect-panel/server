package user

import (
	"github.com/perfect-panel/server/internal/logic/admin/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Reset user subscribe traffic
func ResetUserSubscribeTrafficHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.ResetUserSubscribeTrafficRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := user.NewResetUserSubscribeTrafficLogic(c.Request.Context(), svcCtx)
		err := l.ResetUserSubscribeTraffic(&req)
		result.HttpResult(c, nil, err)
	}
}
