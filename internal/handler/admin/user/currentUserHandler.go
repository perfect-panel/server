package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/admin/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/result"
)

// Current user
func CurrentUserHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		l := user.NewCurrentUserLogic(c.Request.Context(), svcCtx)
		resp, err := l.CurrentUser()
		result.HttpResult(c, resp, err)
	}
}
