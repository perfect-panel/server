package plugin

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/perfect-panel/server/internal/plugin"
	"github.com/perfect-panel/server/internal/svc"
)

// EnableHandler 启用插件 POST /v1/admin/plugin/enable
func EnableHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		name := string(ctx.FormValue("name"))
		if name == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]string{"error": "name is required"})
			return
		}

		mgr, ok := svcCtx.PluginMgr.(*plugin.Manager)
		if !ok || mgr == nil {
			ctx.JSON(consts.StatusServiceUnavailable, map[string]string{"error": "plugin manager not available"})
			return
		}

		if err := mgr.EnablePlugin(name); err != nil {
			ctx.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]string{
			"message": "plugin enabled",
			"name":    name,
		})
	}
}
