package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Delete Current User Account
func DeleteCurrentUserAccountHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		l := user.NewDeleteCurrentUserAccountLogic(c.Request.Context(), svcCtx)
		err := l.DeleteCurrentUserAccount()
		result.HttpResult(c, nil, err)
	}
}
