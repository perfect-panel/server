package fiberserver

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
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

func New(svc *svc.ServiceContext, fallback ...http.Handler) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	registerSubscribeHandlers(app, svc)
	registerServerHandlers(app, svc)
	if len(fallback) > 0 && fallback[0] != nil {
		app.Use(adaptor.HTTPHandler(fallback[0]))
	}
	return app
}

func NewHTTPHandler(svc *svc.ServiceContext, fallback ...http.Handler) http.Handler {
	return adaptor.FiberApp(New(svc, fallback...))
}

func registerSubscribeHandlers(app *fiber.App, svc *svc.ServiceContext) {
	path := svc.Config.Subscribe.SubscribePath
	if path == "" {
		path = "/v1/subscribe/config"
	}
	app.Get(path, subscribeHandler(svc))
	if svc.Config.Subscribe.PanDomain {
		app.Get("/", panDomainSubscribeHandler(svc))
	}
}

func subscribeHandler(svc *svc.ServiceContext) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := requestContext(c)
		req := types.SubscribeRequest{
			Token:  c.Get("token"),
			UA:     c.Get("User-Agent"),
			Flag:   c.Query("flag"),
			Type:   c.Query("type"),
			Params: c.Queries(),
		}
		if req.Token == "" {
			req.Token = c.Query("token")
		}

		if svc.Config.Subscribe.PanDomain {
			domainArr := strings.Split(c.Hostname(), ".")
			if len(domainArr) == 0 {
				return c.Status(fiber.StatusForbidden).SendString("Access denied")
			}
			short, err := tool.FixedUniqueString(req.Token, 8, "")
			if err != nil {
				logger.WithContext(ctx).Errorf("[FiberSubscribeHandler] Generate short token failed: %v", err)
				return c.Status(fiber.StatusInternalServerError).SendString("Internal Server")
			}
			if strings.ToLower(short) != strings.ToLower(domainArr[0]) {
				logger.WithContext(ctx).Debugf("[FiberSubscribeHandler] short token mismatch, short: %s, domain: %s", short, domainArr[0])
				return c.Status(fiber.StatusForbidden).SendString("Access denied")
			}
		}

		if svc.Config.Subscribe.UserAgentLimit && !subscribelogic.IsUserAgentAllowed(ctx, svc, req.UA) {
			return c.Status(fiber.StatusForbidden).SendString("Access denied")
		}
		return buildSubscribeResponse(c, svc, req)
	}
}

func panDomainSubscribeHandler(svc *svc.ServiceContext) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := requestContext(c)
		ua := c.Get("User-Agent")
		if svc.Config.Subscribe.UserAgentLimit && !subscribelogic.IsUserAgentAllowed(ctx, svc, ua) {
			return c.Status(fiber.StatusForbidden).SendString("Access denied")
		}

		domainArr := strings.Split(c.Hostname(), ".")
		if len(domainArr) < 2 {
			return c.Status(fiber.StatusForbidden).SendString("Access denied")
		}
		return buildSubscribeResponse(c, svc, types.SubscribeRequest{
			Token:  domainArr[0],
			Flag:   domainArr[1],
			UA:     ua,
			Params: c.Queries(),
		})
	}
}

func buildSubscribeResponse(c *fiber.Ctx, svc *svc.ServiceContext, req types.SubscribeRequest) error {
	l := subscribelogic.NewSubscribeLogic(requestContext(c), svc, subscribelogic.RequestMeta{
		Host:       c.Hostname(),
		RequestURI: c.OriginalURL(),
		UserAgent:  c.Get("User-Agent"),
		ClientIP:   c.IP(),
	})
	resp, err := l.Handler(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server")
	}
	for key, value := range resp.Headers {
		c.Set(key, value)
	}
	c.Set("subscription-userinfo", resp.Header)
	return c.Status(fiber.StatusOK).Send(resp.Config)
}

func registerServerHandlers(app *fiber.App, svc *svc.ServiceContext) {
	group := app.Group("/v1/server", serverSecretMiddleware(svc))
	group.Get("/config", getServerConfigHandler(svc))
	group.Get("/user", getServerUserListHandler(svc))
}

func serverSecretMiddleware(svc *svc.ServiceContext) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key, ok := c.Queries()["secret_key"]
		if ok && key == svc.Config.Node.NodeSecret {
			return c.Next()
		}
		return c.Status(fiber.StatusForbidden).SendString("Forbidden")
	}
}

func getServerConfigHandler(svc *svc.ServiceContext) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commonReq, err := serverCommonRequest(c)
		if err != nil {
			return writeParamError(c, err)
		}
		req := types.GetServerConfigRequest{ServerCommon: commonReq}
		if validateErr := svc.Validate(&req); validateErr != nil {
			return writeParamError(c, validateErr)
		}

		l := serverlogic.NewGetServerConfigLogic(requestContext(c), svc, serverlogic.RequestMeta{
			IfNoneMatch: c.Get("If-None-Match"),
		})
		resp, err := l.GetServerConfig(&req)
		writeHeaders(c, l.ResponseMeta().Headers)
		if err != nil {
			if errors.Is(err, xerr.StatusNotModified) {
				return c.Status(fiber.StatusNotModified).SendString("Not Modified")
			}
			return c.Status(fiber.StatusNotFound).SendString("Not Found")
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

func getServerUserListHandler(svc *svc.ServiceContext) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commonReq, err := serverCommonRequest(c)
		if err != nil {
			return writeParamError(c, err)
		}
		req := types.GetServerUserListRequest{ServerCommon: commonReq}
		if validateErr := svc.Validate(&req); validateErr != nil {
			return writeParamError(c, validateErr)
		}

		l := serverlogic.NewGetServerUserListLogic(requestContext(c), svc, serverlogic.RequestMeta{
			IfNoneMatch: c.Get("If-None-Match"),
		})
		resp, err := l.GetServerUserList(&req)
		writeHeaders(c, l.ResponseMeta().Headers)
		if err != nil {
			if errors.Is(err, xerr.StatusNotModified) {
				return c.Status(fiber.StatusNotModified).SendString("Not Modified")
			}
			return c.Status(fiber.StatusNotFound).SendString("Not Found")
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}
}

func serverCommonRequest(c *fiber.Ctx) (types.ServerCommon, error) {
	var serverID int64
	if rawServerID := c.Query("server_id"); rawServerID != "" {
		id, err := strconv.ParseInt(rawServerID, 10, 64)
		if err != nil {
			return types.ServerCommon{}, err
		}
		serverID = id
	}
	return types.ServerCommon{
		Protocol:  c.Query("protocol"),
		ServerId:  serverID,
		SecretKey: c.Query("secret_key"),
	}, nil
}

func writeHeaders(c *fiber.Ctx, headers map[string]string) {
	for key, value := range headers {
		c.Set(key, value)
	}
}

func writeParamError(c *fiber.Ctx, err error) error {
	resp := result.BuildParamErrorResult(err)
	return c.Status(resp.StatusCode).JSON(resp.Body)
}

func requestContext(c *fiber.Ctx) context.Context {
	ctx := c.UserContext()
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
