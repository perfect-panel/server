package common

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/common"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get Privacy Policy
func GetPrivacyPolicyHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := common.NewGetPrivacyPolicyLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetPrivacyPolicy()
		result.HttpResult(c, resp, err)
	}
}
