package hertzserver

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/perfect-panel/server/internal/svc"
)

func TestNewFallbackHandler(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	addr := ln.Addr().String()

	app := newServer(&svc.ServiceContext{}, []config.Option{
		server.WithListener(ln),
		server.WithTransport(standard.NewTransporter),
		server.WithDisablePrintRoute(true),
	}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Fallback", "gin")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("fallback"))
	}))

	go app.Start()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := app.Shutdown(ctx); err != nil {
			t.Fatalf("shutdown failed: %v", err)
		}
	}()

	deadline := time.Now().Add(time.Second)
	for !app.Engine().IsRunning() {
		if time.Now().After(deadline) {
			t.Fatal("server did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}

	resp, err := http.Get("http://" + addr + "/fallback")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if got := resp.StatusCode; got != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, got)
	}
	if got := resp.Header.Get("X-Fallback"); got != "gin" {
		t.Fatalf("expected fallback header, got %q", got)
	}
	if got := string(body); got != "fallback" {
		t.Fatalf("expected fallback body, got %q", got)
	}
}
