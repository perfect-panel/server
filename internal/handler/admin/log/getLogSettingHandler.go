package log

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/log"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get log setting
func GetLogSettingHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := log.NewGetLogSettingLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetLogSetting()
		result.HttpResult(c, resp, err)
	}
}
