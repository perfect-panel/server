package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/perfect-panel/ppanel-server/pkg/constant"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/perfect-panel/ppanel-server/internal/svc"
	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/trace"
)

// statusByWriter returns a span status code and message for an HTTP status code
// value returned by a server. Status codes in the 400-499 range are not
// returned as errors.
func statusByWriter(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}

func requestAttributes(req *http.Request) []attribute.KeyValue {
	protoN := strings.SplitN(req.Proto, "/", 2)
	remoteAddrN := strings.SplitN(req.RemoteAddr, ":", 2)

	return []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String(req.Method),
		semconv.HTTPUserAgentKey.String(req.UserAgent()),
		semconv.HTTPRequestContentLengthKey.Int64(req.ContentLength),

		semconv.URLFullKey.String(req.URL.String()),
		semconv.URLSchemeKey.String(req.URL.Scheme),
		semconv.URLFragmentKey.String(req.URL.Fragment),
		semconv.URLPathKey.String(req.URL.Path),
		semconv.URLQueryKey.String(req.URL.RawQuery),

		semconv.NetworkProtocolNameKey.String(strings.ToLower(protoN[0])),
		semconv.NetworkProtocolVersionKey.String(protoN[1]),

		semconv.ClientAddressKey.String(remoteAddrN[0]),
		semconv.ClientPortKey.String(remoteAddrN[1]),
	}
}

func TraceMiddleware(_ *svc.ServiceContext) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		tracer := trace.TracerFromContext(ctx)

		spanName := c.FullPath()
		method := c.Request.Method

		ctx, span := tracer.Start(
			ctx,
			fmt.Sprintf("%s %s", method, spanName),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		)
		defer span.End()

		requestId, err := uuid.NewV7()
		if err != nil {
			logger.Errorw(
				"failed to generate request id in uuid v7 format, fallback to uuid v4",
				logger.Field("error", err),
			)
			requestId = uuid.New()
		}
		c.Header(trace.RequestIdKey, requestId.String())

		span.SetAttributes(requestAttributes(c.Request)...)
		span.SetAttributes(
			attribute.String("http.request_id", requestId.String()),
			semconv.HTTPRouteKey.String(c.FullPath()),
		)
		// context with request host
		ctx = context.WithValue(ctx, constant.CtxKeyRequestHost, c.Request.Host)
		// restructure context
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// handle response related attributes
		status := c.Writer.Status()
		span.SetStatus(statusByWriter(status))
		if status > 0 {
			span.SetAttributes(semconv.HTTPResponseStatusCodeKey.Int(status))
		}
		if len(c.Errors) > 0 {
			span.SetStatus(codes.Error, c.Errors.String())
			for _, err := range c.Errors {
				span.RecordError(err.Err)
			}
		}

		span.SetAttributes(semconv.HTTPResponseBodySizeKey.Int(c.Writer.Size()))
	}
}
