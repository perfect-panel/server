package announcement

import (
	"github.com/perfect-panel/server/internal/logic/admin/announcement"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

// Create announcement
func CreateAnnouncementHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.CreateAnnouncementRequest
		_ = c.ShouldBind(&req)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := announcement.NewCreateAnnouncementLogic(c.Request.Context(), svcCtx)
		err := l.CreateAnnouncement(&req)
		result.HttpResult(c, nil, err)
	}
}
