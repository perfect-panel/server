package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/user"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Update User Subscribe Note
func UpdateUserSubscribeNoteHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.UpdateUserSubscribeNoteRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := user.NewUpdateUserSubscribeNoteLogic(c.Request.Context(), svcCtx)
		err := l.UpdateUserSubscribeNote(&req)
		result.HttpResult(c, nil, err)
	}
}
