package common

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/common"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Get Subscription
func GetSubscriptionHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := common.NewGetSubscriptionLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetSubscription()
		result.HttpResult(c, resp, err)
	}
}
