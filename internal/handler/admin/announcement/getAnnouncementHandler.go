package announcement

import (
	"github.com/perfect-panel/server/internal/logic/admin/announcement"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Get announcement
func GetAnnouncementHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.GetAnnouncementRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := announcement.NewGetAnnouncementLogic(c.Request.Context(), svcCtx)
		resp, err := l.GetAnnouncement(&req)
		result.HttpResult(c, resp, err)
	}
}
