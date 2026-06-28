package server

import (
	"github.com/perfect-panel/server/internal/logic/admin/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Update Server
func UpdateServerHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.UpdateServerRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := server.NewUpdateServerLogic(c.Request.Context(), svcCtx)
		err := l.UpdateServer(&req)
		result.HttpResult(c, nil, err)
	}
}
