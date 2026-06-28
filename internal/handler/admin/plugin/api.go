package plugin

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	runtimeplugin "github.com/perfect-panel/server/internal/plugin"
	"github.com/perfect-panel/server/internal/svc"
)

type apiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type pluginListResponse struct {
	List  []runtimeplugin.PluginInfo `json:"list"`
	Total int                        `json:"total"`
}

type pluginActionResponse struct {
	Name    string `json:"name"`
	Action  string `json:"action"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func writeOK(ctx *app.RequestContext, data interface{}) {
	ctx.JSON(consts.StatusOK, apiResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

func writeError(ctx *app.RequestContext, code int, message string) {
	ctx.JSON(consts.StatusOK, apiResponse{
		Code:    code,
		Message: message,
	})
}

func pluginManager(svcCtx *svc.ServiceContext, ctx *app.RequestContext) (*runtimeplugin.Manager, bool) {
	mgr, ok := svcCtx.PluginMgr.(*runtimeplugin.Manager)
	if !ok || mgr == nil {
		writeError(ctx, consts.StatusServiceUnavailable, "plugin manager not available")
		return nil, false
	}
	return mgr, true
}

func pluginName(ctx *app.RequestContext) (string, bool) {
	name := strings.TrimSpace(ctx.Param("name"))
	if err := runtimeplugin.ValidatePluginName(name); err != nil {
		writeError(ctx, consts.StatusBadRequest, err.Error())
		return "", false
	}
	return name, true
}

// InstalledUploadHandler 上传并安装插件 zip 包 POST /v1/admin/plugins/upload
func InstalledUploadHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}

		fileHeader, err := ctx.FormFile("file")
		if err != nil {
			writeError(ctx, consts.StatusBadRequest, "plugin package file is required")
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			writeError(ctx, consts.StatusBadRequest, fmt.Sprintf("open plugin package: %v", err))
			return
		}
		defer file.Close()

		result, err := mgr.InstallPluginArchive(c, file, runtimeplugin.PluginInstallOptions{
			Replace: parseBoolForm(ctx, "replace"),
			Enable:  parseBoolForm(ctx, "enable"),
		})
		if err != nil {
			writeError(ctx, consts.StatusBadRequest, err.Error())
			return
		}
		writeOK(ctx, result)
	}
}

// InstalledListHandler 获取已安装插件列表 GET /v1/admin/plugins
func InstalledListHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}

		q := strings.ToLower(strings.TrimSpace(ctx.Query("q")))
		status := strings.TrimSpace(ctx.Query("status"))
		page, size := pagination(ctx)

		all := mgr.ListInstalledPlugins()
		filtered := make([]runtimeplugin.PluginInfo, 0, len(all))
		for _, item := range all {
			if status != "" && string(item.Status) != status {
				continue
			}
			if q != "" {
				haystack := strings.ToLower(item.Name + " " + item.Description + " " + item.Author)
				if !strings.Contains(haystack, q) {
					continue
				}
			}
			filtered = append(filtered, item)
		}

		total := len(filtered)
		start := (page - 1) * size
		if start > total {
			start = total
		}
		end := start + size
		if end > total {
			end = total
		}

		writeOK(ctx, pluginListResponse{
			List:  filtered[start:end],
			Total: total,
		})
	}
}

// InstalledReloadAllHandler 重新扫描并加载所有插件 POST /v1/admin/plugins/reload-all
func InstalledReloadAllHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}
		plugins := mgr.ReloadAllPlugins(c)
		writeOK(ctx, pluginListResponse{
			List:  plugins,
			Total: len(plugins),
		})
	}
}

// InstalledDetailHandler 获取插件详情 GET /v1/admin/plugins/:name
func InstalledDetailHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}
		name, ok := pluginName(ctx)
		if !ok {
			return
		}
		info, exists := mgr.GetInstalledPluginInfo(name)
		if !exists {
			writeError(ctx, consts.StatusNotFound, "plugin not found")
			return
		}
		writeOK(ctx, info)
	}
}

// InstalledValidateHandler 校验插件目录和 WASM POST /v1/admin/plugins/:name/validate
func InstalledValidateHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}
		name, ok := pluginName(ctx)
		if !ok {
			return
		}
		writeOK(ctx, mgr.ValidateInstalledPlugin(name))
	}
}

// InstalledManifestHandler 获取插件清单 GET /v1/admin/plugins/:name/manifest
func InstalledManifestHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}
		name, ok := pluginName(ctx)
		if !ok {
			return
		}
		manifest, err := mgr.GetInstalledManifest(name)
		if err != nil {
			writeError(ctx, consts.StatusNotFound, err.Error())
			return
		}
		writeOK(ctx, manifest)
	}
}

// InstalledRoutesHandler 获取插件运行时路由 GET /v1/admin/plugins/:name/routes
func InstalledRoutesHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}
		name, ok := pluginName(ctx)
		if !ok {
			return
		}
		writeOK(ctx, mgr.ListPluginRoutes(name))
	}
}

// InstalledMiddlewareHandler 获取插件运行时中间件 GET /v1/admin/plugins/:name/middlewares
func InstalledMiddlewareHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}
		name, ok := pluginName(ctx)
		if !ok {
			return
		}
		writeOK(ctx, mgr.ListPluginMiddleware(name))
	}
}

// InstalledEventsHandler 获取插件事件订阅 GET /v1/admin/plugins/:name/events
func InstalledEventsHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}
		name, ok := pluginName(ctx)
		if !ok {
			return
		}
		writeOK(ctx, mgr.ListPluginEvents(name))
	}
}

// InstalledHealthHandler 获取插件健康状态 GET /v1/admin/plugins/:name/health
func InstalledHealthHandler(svcCtx *svc.ServiceContext) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}
		name, ok := pluginName(ctx)
		if !ok {
			return
		}
		health, exists := mgr.GetPluginHealth(name)
		if !exists {
			writeError(ctx, consts.StatusNotFound, "plugin not found")
			return
		}
		writeOK(ctx, health)
	}
}

// InstalledActionHandler 执行插件生命周期动作。
func InstalledActionHandler(svcCtx *svc.ServiceContext, action string) app.HandlerFunc {
	return func(_ context.Context, ctx *app.RequestContext) {
		mgr, ok := pluginManager(svcCtx, ctx)
		if !ok {
			return
		}
		name, ok := pluginName(ctx)
		if !ok {
			return
		}

		var err error
		switch action {
		case "enable":
			err = mgr.EnablePlugin(name)
		case "disable":
			err = mgr.DisablePlugin(name)
		case "reload", "restart":
			err = mgr.ReloadPlugin(name)
		default:
			writeError(ctx, consts.StatusBadRequest, "unsupported plugin action")
			return
		}
		if err != nil {
			writeError(ctx, consts.StatusInternalServerError, err.Error())
			return
		}

		status := string(runtimeplugin.StatusUnloaded)
		if info, exists := mgr.GetInstalledPluginInfo(name); exists {
			status = string(info.Status)
		}
		writeOK(ctx, pluginActionResponse{
			Name:    name,
			Action:  action,
			Status:  status,
			Message: "plugin " + action + " completed",
		})
	}
}

func pagination(ctx *app.RequestContext) (int, int) {
	page, _ := strconv.Atoi(ctx.Query("page"))
	size, _ := strconv.Atoi(ctx.Query("size"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func parseBoolForm(ctx *app.RequestContext, key string) bool {
	switch strings.ToLower(strings.TrimSpace(ctx.PostForm(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
