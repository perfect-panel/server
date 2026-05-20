package console

import (
	"github.com/perfect-panel/server/internal/logic/admin/console"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Query user statistics
func QueryUserStatisticsHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := console.NewQueryUserStatisticsLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryUserStatistics()
		result.HttpResult(c, resp, err)
	}
}
