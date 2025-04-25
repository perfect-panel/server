package common

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/ppanel-server/internal/logic/common"
	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/result"
)

// Get stat
func GetStatHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := common.NewGetStatLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetStat()
		result.HttpResult(c, resp, err)
	}
}
