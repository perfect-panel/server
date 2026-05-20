package hertzserver

import (
	"context"
	"net/http"
	"testing"

	appconfig "github.com/perfect-panel/server/internal/config"
	"github.com/perfect-panel/server/internal/svc"
)

func TestServerSecretMiddlewareBlocksMigratedPost(t *testing.T) {
	app := newTestServer("secret")

	status, body := performNativeRequest(app, http.MethodPost, "/v1/server/online?secret_key=wrong")
	if status != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, status)
	}
	if body != "Forbidden" {
		t.Fatalf("expected forbidden body, got %q", body)
	}
}

func TestQueryServerProtocolConfigRejectsInvalidID(t *testing.T) {
	app := newTestServer("secret")

	status, body := performNativeRequest(app, http.MethodGet, "/v2/server/not-a-number?secret_key=secret")
	if status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, status)
	}
	if body != "Invalid Params" {
		t.Fatalf("expected invalid params body, got %q", body)
	}
}

func TestQueryServerProtocolConfigRejectsInvalidSecret(t *testing.T) {
	app := newTestServer("secret")

	status, body := performNativeRequest(app, http.MethodGet, "/v2/server/1?secret_key=wrong")
	if status != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, status)
	}
	if body != "Unauthorized" {
		t.Fatalf("expected unauthorized body, got %q", body)
	}
}

func newTestServer(secret string) *Server {
	return New(&svc.ServiceContext{
		Config: appconfig.Config{
			Node: appconfig.NodeConfig{
				NodeSecret: secret,
			},
		},
	}, "127.0.0.1:0", nil)
}

func performNativeRequest(server *Server, method, uri string) (int, string) {
	ctx := server.Engine().NewContext()
	ctx.Request.SetRequestURI(uri)
	ctx.Request.Header.SetMethod(method)
	server.Engine().ServeHTTP(context.Background(), ctx)
	return ctx.Response.StatusCode(), string(ctx.Response.Body())
}
