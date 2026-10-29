//go:build integration

package assistant

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nostosx/agentx-examples/infrastructure/config"
	serverinfra "github.com/nostosx/agentx-examples/infrastructure/server"
	"github.com/nostosx/agentx/model"
	lifecycle "github.com/nostosx/servex/app"
	"github.com/nostosx/servex/health"
)

func TestProtocolRoutes(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	m := model.Func(func(context.Context, *model.Request) (*model.Response, error) {
		calls.Add(1)
		return &model.Response{Message: model.AssistantText("fixture-answer")}, nil
	})
	token := strings.Repeat("a", 32)
	service, err := New(m, token, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var app *lifecycle.App
	servers, err := serverinfra.New(ctx, config.Config{Address: "127.0.0.1:0", Token: token}, health.CheckerFunc(func(ctx context.Context) error { return app.Check(ctx) }), nil)
	if err != nil {
		t.Fatal(err)
	}
	service.Register(ctx, servers.HTTP)
	app = lifecycle.New(lifecycle.WithHooks(servers.HTTP))
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		stop, done := context.WithTimeout(context.Background(), time.Second)
		defer done()
		if err := app.Stop(stop); err != nil {
			t.Error(err)
		}
	})
	client := &http.Client{Timeout: 3 * time.Second}
	defer client.CloseIdleConnections()
	for _, tt := range []struct {
		name, path, body, token string
		status                  int
	}{
		{name: "health", path: "/readyz", status: http.StatusOK},
		{name: "unauthenticated session", path: "/sessions/demo", status: http.StatusUnauthorized},
		{name: "wrong token", path: "/sessions/demo", token: "wrong", status: http.StatusUnauthorized},
		{name: "run", path: "/sessions/demo/runs", body: `{"input":"hello"}`, token: token, status: http.StatusOK},
		{name: "session", path: "/sessions/demo", token: token, status: http.StatusOK},
	} {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodGet
			if tt.body != "" {
				method = http.MethodPost
			}
			req, err := http.NewRequestWithContext(ctx, method, servers.HTTP.URL().String()+tt.path, strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			req.Header.Set("Content-Type", "application/json")
			response, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			data, readErr := io.ReadAll(response.Body)
			closeErr := response.Body.Close()
			if readErr != nil || closeErr != nil || response.StatusCode != tt.status {
				t.Fatalf("status=%d body=%s read=%v close=%v", response.StatusCode, data, readErr, closeErr)
			}
			if tt.name == "run" && !strings.Contains(string(data), "fixture-answer") {
				t.Fatalf("missing model output: %s", data)
			}
		})
	}
	if calls.Load() != 1 {
		t.Fatalf("model calls=%d", calls.Load())
	}
}

func TestShutdownCancelsModelRequest(t *testing.T) {
	t.Parallel()
	entered, canceled := make(chan struct{}), make(chan struct{})
	m := model.Func(func(ctx context.Context, _ *model.Request) (*model.Response, error) {
		close(entered)
		<-ctx.Done()
		close(canceled)
		return nil, ctx.Err()
	})
	token := strings.Repeat("a", 32)
	service, err := New(m, token, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var app *lifecycle.App
	servers, err := serverinfra.New(ctx, config.Config{Address: "127.0.0.1:0", Token: token}, health.CheckerFunc(func(ctx context.Context) error { return app.Check(ctx) }), nil)
	if err != nil {
		t.Fatal(err)
	}
	service.Register(ctx, servers.HTTP)
	app = lifecycle.New(lifecycle.WithHooks(servers.HTTP))
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		stop, done := context.WithTimeout(context.Background(), time.Second)
		defer done()
		if err := app.Stop(stop); err != nil {
			t.Error(err)
		}
	})
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, servers.HTTP.URL().String()+"/sessions/demo/runs", strings.NewReader(`{"input":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	done := make(chan error, 1)
	go func() {
		response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
		if err == nil {
			_, err = io.Copy(io.Discard, response.Body)
			errClose := response.Body.Close()
			if err == nil {
				err = errClose
			}
		}
		done <- err
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("model did not start")
	}
	cancel()
	select {
	case <-canceled:
	case <-time.After(3 * time.Second):
		t.Fatal("model did not cancel")
	}
	stop, release := context.WithTimeout(context.Background(), 3*time.Second)
	defer release()
	if err := app.Stop(stop); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("request did not finish")
	}
}
