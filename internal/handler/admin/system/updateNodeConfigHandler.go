package system

import (
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Update node config
func UpdateNodeConfigHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.NodeConfig
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := system.NewUpdateNodeConfigLogic(c.Request.Context(), svcCtx)
		err := l.UpdateNodeConfig(&req)
		result.HttpResult(c, nil, err)
	}
}
