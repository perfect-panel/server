package hertzserver

import (
	"context"
	"crypto/tls"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	serverlogic "github.com/perfect-panel/server/internal/logic/server"
	subscribelogic "github.com/perfect-panel/server/internal/logic/subscribe"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/result"
	"github.com/perfect-panel/server/pkg/tool"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type Server struct {
	h *server.Hertz
}

func New(svc *svc.ServiceContext, addr string, tlsConfig *tls.Config, fallback ...http.Handler) *Server {
	opts := []config.Option{
		server.WithHostPorts(addr),
		server.WithDisablePrintRoute(true),
	}
	if tlsConfig != nil {
		opts = append(opts, server.WithTLS(tlsConfig))
	}

	return newServer(svc, opts, fallback...)
}

func newServer(svc *svc.ServiceContext, opts []config.Option, fallback ...http.Handler) *Server {
	h := server.Default(opts...)
	registerSubscribeHandlers(h, svc)
	registerServerHandlers(h, svc)
	if len(fallback) > 0 && fallback[0] != nil {
		h.NoRoute(adaptor.HertzHandler(fallback[0]))
	}
	return &Server{h: h}
}

func (s *Server) Start() {
	if err := s.h.Run(); err != nil {
		logger.Errorf("server start error: %s", err.Error())
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.h.Shutdown(ctx)
}

func (s *Server) Engine() *server.Hertz {
	return s.h
}

func registerSubscribeHandlers(h *server.Hertz, svc *svc.ServiceContext) {
	path := svc.Config.Subscribe.SubscribePath
	if path == "" {
		path = "/v1/subscribe/config"
	}
	h.GET(path, subscribeHandler(svc))
	if svc.Config.Subscribe.PanDomain {
		h.GET("/", panDomainSubscribeHandler(svc))
	}
}

func subscribeHandler(svc *svc.ServiceContext) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		req := types.SubscribeRequest{
			Token:  string(ctx.GetHeader("token")),
			UA:     string(ctx.UserAgent()),
			Flag:   ctx.Query("flag"),
			Type:   ctx.Query("type"),
			Params: queryMap(ctx),
		}
		if req.Token == "" {
			req.Token = ctx.Query("token")
		}

		if svc.Config.Subscribe.PanDomain {
			domainArr := strings.Split(string(ctx.Host()), ".")
			if len(domainArr) == 0 {
				ctx.String(consts.StatusForbidden, "Access denied")
				return
			}
			short, err := tool.FixedUniqueString(req.Token, 8, "")
			if err != nil {
				logger.WithContext(c).Errorf("[HertzSubscribeHandler] Generate short token failed: %v", err)
				ctx.String(consts.StatusInternalServerError, "Internal Server")
				return
			}
			if strings.ToLower(short) != strings.ToLower(domainArr[0]) {
				logger.WithContext(c).Debugf("[HertzSubscribeHandler] short token mismatch, short: %s, domain: %s", short, domainArr[0])
				ctx.String(consts.StatusForbidden, "Access denied")
				return
			}
		}

		if svc.Config.Subscribe.UserAgentLimit && !subscribelogic.IsUserAgentAllowed(c, svc, req.UA) {
			ctx.String(consts.StatusForbidden, "Access denied")
			return
		}
		buildSubscribeResponse(c, ctx, svc, req)
	}
}

func panDomainSubscribeHandler(svc *svc.ServiceContext) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ua := string(ctx.UserAgent())
		if svc.Config.Subscribe.UserAgentLimit && !subscribelogic.IsUserAgentAllowed(c, svc, ua) {
			ctx.String(consts.StatusForbidden, "Access denied")
			return
		}

		domainArr := strings.Split(string(ctx.Host()), ".")
		if len(domainArr) < 2 {
			ctx.String(consts.StatusForbidden, "Access denied")
			return
		}
		buildSubscribeResponse(c, ctx, svc, types.SubscribeRequest{
			Token:  domainArr[0],
			Flag:   domainArr[1],
			UA:     ua,
			Params: queryMap(ctx),
		})
	}
}

func buildSubscribeResponse(c context.Context, ctx *app.RequestContext, svc *svc.ServiceContext, req types.SubscribeRequest) {
	l := subscribelogic.NewSubscribeLogic(c, svc, subscribelogic.RequestMeta{
		Host:       string(ctx.Host()),
		RequestURI: string(ctx.URI().RequestURI()),
		UserAgent:  string(ctx.UserAgent()),
		ClientIP:   ctx.ClientIP(),
	})
	resp, err := l.Handler(&req)
	if err != nil {
		ctx.String(consts.StatusInternalServerError, "Internal Server")
		return
	}
	writeHeaders(ctx, resp.Headers)
	ctx.Header("subscription-userinfo", resp.Header)
	ctx.Data(consts.StatusOK, "text/plain; charset=utf-8", resp.Config)
}

func registerServerHandlers(h *server.Hertz, svc *svc.ServiceContext) {
	group := h.Group("/v1/server", serverSecretMiddleware(svc))
	group.GET("/config", getServerConfigHandler(svc))
	group.GET("/user", getServerUserListHandler(svc))
}

func serverSecretMiddleware(svc *svc.ServiceContext) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		key, ok := ctx.GetQuery("secret_key")
		if ok && key == svc.Config.Node.NodeSecret {
			ctx.Next(c)
			return
		}
		ctx.String(consts.StatusForbidden, "Forbidden")
		ctx.Abort()
	}
}

func getServerConfigHandler(svc *svc.ServiceContext) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		commonReq, err := serverCommonRequest(ctx)
		if err != nil {
			writeParamError(ctx, err)
			return
		}
		req := types.GetServerConfigRequest{ServerCommon: commonReq}
		if validateErr := svc.Validate(&req); validateErr != nil {
			writeParamError(ctx, validateErr)
			return
		}

		l := serverlogic.NewGetServerConfigLogic(c, svc, serverlogic.RequestMeta{
			IfNoneMatch: string(ctx.GetHeader("If-None-Match")),
		})
		resp, err := l.GetServerConfig(&req)
		writeHeaders(ctx, l.ResponseMeta().Headers)
		if err != nil {
			if errors.Is(err, xerr.StatusNotModified) {
				ctx.String(consts.StatusNotModified, "Not Modified")
				return
			}
			ctx.String(consts.StatusNotFound, "Not Found")
			return
		}
		ctx.JSON(consts.StatusOK, resp)
	}
}

func getServerUserListHandler(svc *svc.ServiceContext) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		commonReq, err := serverCommonRequest(ctx)
		if err != nil {
			writeParamError(ctx, err)
			return
		}
		req := types.GetServerUserListRequest{ServerCommon: commonReq}
		if validateErr := svc.Validate(&req); validateErr != nil {
			writeParamError(ctx, validateErr)
			return
		}

		l := serverlogic.NewGetServerUserListLogic(c, svc, serverlogic.RequestMeta{
			IfNoneMatch: string(ctx.GetHeader("If-None-Match")),
		})
		resp, err := l.GetServerUserList(&req)
		writeHeaders(ctx, l.ResponseMeta().Headers)
		if err != nil {
			if errors.Is(err, xerr.StatusNotModified) {
				ctx.String(consts.StatusNotModified, "Not Modified")
				return
			}
			ctx.String(consts.StatusNotFound, "Not Found")
			return
		}
		ctx.JSON(consts.StatusOK, resp)
	}
}

func serverCommonRequest(ctx *app.RequestContext) (types.ServerCommon, error) {
	var serverID int64
	if rawServerID := ctx.Query("server_id"); rawServerID != "" {
		id, err := strconv.ParseInt(rawServerID, 10, 64)
		if err != nil {
			return types.ServerCommon{}, err
		}
		serverID = id
	}
	return types.ServerCommon{
		Protocol:  ctx.Query("protocol"),
		ServerId:  serverID,
		SecretKey: ctx.Query("secret_key"),
	}, nil
}

func queryMap(ctx *app.RequestContext) map[string]string {
	params := make(map[string]string)
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		params[string(key)] = string(value)
	})
	return params
}

func writeHeaders(ctx *app.RequestContext, headers map[string]string) {
	for key, value := range headers {
		ctx.Header(key, value)
	}
}

func writeParamError(ctx *app.RequestContext, err error) {
	resp := result.BuildParamErrorResult(err)
	ctx.JSON(resp.StatusCode, resp.Body)
}
