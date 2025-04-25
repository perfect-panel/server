package common

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/common"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Get sms verification code
func SendSmsCodeHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.SendSmsCodeRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := common.NewSendSmsCodeLogic(c.Request.Context(), svcCtx)
		resp, err := l.SendSmsCode(&req)
		result.HttpResult(c, resp, err)
	}
}
