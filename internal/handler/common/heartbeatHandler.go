package common

import (
	"github.com/perfect-panel/server/internal/logic/common"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Heartbeat
func HeartbeatHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := common.NewHeartbeatLogic(c.Request.Context(), svcCtx)
		resp, err := l.Heartbeat()
		result.HttpResult(c, resp, err)
	}
}
