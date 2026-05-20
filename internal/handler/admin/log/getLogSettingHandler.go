package log

import (
	"github.com/perfect-panel/server/internal/logic/admin/log"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get log setting
func GetLogSettingHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {

		l := log.NewGetLogSettingLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetLogSetting()
		result.HttpResult(c, resp, err)
	}
}
