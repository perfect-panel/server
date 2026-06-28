package common

import (
	"github.com/perfect-panel/server/internal/logic/common"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get stat
func GetStatHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := common.NewGetStatLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetStat()
		result.HttpResult(c, resp, err)
	}
}
