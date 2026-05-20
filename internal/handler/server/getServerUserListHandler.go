package server

import (
	"github.com/perfect-panel/server/internal/logic/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

// Get user list
func GetServerUserListHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.GetServerUserListRequest
		_ = c.ShouldBind(&req)
		_ = c.ShouldBindQuery(&req.ServerCommon)
		validateErr := svcCtx.Validate(&req)
		if validateErr != nil {
			result.ParamErrorResult(c, validateErr)
			return
		}

		l := server.NewGetServerUserListLogic(c.Request.Context(), svcCtx, server.RequestMeta{
			IfNoneMatch: c.GetHeader("If-None-Match"),
		})
		resp, err := l.GetServerUserList(&req)
		for key, value := range l.ResponseMeta().Headers {
			c.Header(key, value)
		}
		if err != nil {
			if errors.Is(err, xerr.StatusNotModified) {
				c.String(304, "Not Modified")
				return
			}
			c.String(404, "Not Found")
			return
		}
		c.JSON(200, resp)
	}
}
