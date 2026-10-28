//go:build integration

package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/nostosx/servex/app"
	"github.com/nostosx/servex/health"
)

func TestListenersAndGracefulShutdown(t *testing.T) {
	cfg := testConfig()
	var application *app.App
	check := health.CheckerFunc(func(ctx context.Context) error { return application.Check(ctx) })
	servers, err := New(t.Context(), cfg, check)
	if err != nil {
		t.Fatal(err)
	}
	application = app.New(app.WithHooks(servers.HTTP), app.WithStopTimeout(10*time.Second))
	ctx, stop := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- application.Run(ctx) }()
	t.Cleanup(func() {
		stop()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("shutdown: %v", err)
			}
		case <-time.After(15 * time.Second):
			t.Error("shutdown timed out")
		}
	})
	deadline := time.Now().Add(10 * time.Second)
	for servers.HTTP.URL() == nil {
		if time.Now().After(deadline) {
			t.Fatal("listeners did not start")
		}
		time.Sleep(time.Millisecond)
	}
	client := &http.Client{Timeout: 2 * time.Second}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, servers.HTTP.URL().String()+"/readyz", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_, readErr := io.Copy(io.Discard, response.Body)
	closeErr := response.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("readiness=%d", response.StatusCode)
	}
	request, err = http.NewRequestWithContext(t.Context(), http.MethodGet, servers.HTTP.URL().String()+"/v1/unconfigured", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err = client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected business endpoint: %d", response.StatusCode)
	}
}
