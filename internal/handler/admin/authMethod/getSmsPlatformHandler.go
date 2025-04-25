package authMethod

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/authMethod"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get sms support platform
func GetSmsPlatformHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := authMethod.NewGetSmsPlatformLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetSmsPlatform()
		result.HttpResult(c, resp, err)
	}
}
