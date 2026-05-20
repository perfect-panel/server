package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func CorsMiddleware(c context.Context, ctx *app.RequestContext) {
	origin := string(ctx.GetHeader("Origin"))
	if origin != "" {
		ctx.Header("Access-Control-Allow-Origin", origin)
	} else {
		ctx.Header("Access-Control-Allow-Origin", "*")
	}
	ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Origin, X-CSRF-Token, Authorization, AccessToken, Token, Range")
	ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
	ctx.Header("Access-Control-Allow-Credentials", "true")
	ctx.Header("Access-Control-Max-Age", "172800")
	if string(ctx.Method()) == consts.MethodOptions {
		ctx.AbortWithStatus(consts.StatusNoContent)
		return
	}

	ctx.Next(c)
}
