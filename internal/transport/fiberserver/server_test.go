package fiberserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/perfect-panel/server/internal/svc"
)

func TestNewFallbackHandler(t *testing.T) {
	app := New(&svc.ServiceContext{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Fallback", "gin")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("fallback"))
	}))

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/fallback", nil))
	if err != nil {
		t.Fatalf("fiber test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, resp.StatusCode)
	}
	if resp.Header.Get("X-Fallback") != "gin" {
		t.Fatalf("expected fallback header, got %q", resp.Header.Get("X-Fallback"))
	}
}
