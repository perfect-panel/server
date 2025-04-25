package system

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// get Privacy Policy Config
func GetPrivacyPolicyConfigHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := system.NewGetPrivacyPolicyConfigLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetPrivacyPolicyConfig()
		result.HttpResult(c, resp, err)
	}
}
