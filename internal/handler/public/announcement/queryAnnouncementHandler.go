package announcement

import (
	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/logic/public/announcement"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/result"
)

// Query announcement
func QueryAnnouncementHandler(svcCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req types.QueryAnnouncementRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := announcement.NewQueryAnnouncementLogic(c.Request.Context(), svcCtx)
		resp, err := l.QueryAnnouncement(&req)
		result.HttpResult(c, resp, err)
	}
}
