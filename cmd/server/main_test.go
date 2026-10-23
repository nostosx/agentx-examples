package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestValidateAddress(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, addr string
		valid      bool
	}{
		{name: "IPv4", addr: "127.0.0.1:8080", valid: true},
		{name: "IPv6", addr: "[::1]:8080", valid: true},
		{name: "localhost", addr: "localhost:8080", valid: true},
		{name: "wildcard", addr: ":8080"},
		{name: "all IPv4", addr: "0.0.0.0:8080"},
		{name: "all IPv6", addr: "[::]:8080"},
		{name: "remote IP", addr: "192.0.2.1:8080"},
		{name: "DNS name", addr: "example.com:8080"},
		{name: "malformed", addr: "localhost"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateAddress(tt.addr); (err == nil) != tt.valid {
				t.Fatalf("validateAddress(%q) = %v", tt.addr, err)
			}
		})
	}
}

func TestRequireToken(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		query      string
		tokens     []string
		authorized bool
	}{
		{name: "missing"},
		{name: "wrong", tokens: []string{"Bearer wrong"}},
		{name: "query token", query: "?token=secret"},
		{name: "missing scheme", tokens: []string{"secret"}},
		{name: "wrong scheme", tokens: []string{"Basic secret"}},
		{name: "duplicate", tokens: []string{"Bearer secret", "Bearer secret"}},
		{name: "valid", tokens: []string{"Bearer secret"}, authorized: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			handler := requireToken("secret", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				w.WriteHeader(http.StatusAccepted)
			}))
			req := httptest.NewRequest(http.MethodPost, "/sessions/demo/resume"+tt.query, nil)
			for _, token := range tt.tokens {
				req.Header.Add("Authorization", token)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if tt.authorized {
				if response.Code != http.StatusAccepted || calls != 1 {
					t.Fatalf("response=%d calls=%d", response.Code, calls)
				}
				return
			}
			if response.Code != http.StatusUnauthorized || calls != 0 {
				t.Fatalf("response=%d calls=%d", response.Code, calls)
			}
			if got := response.Header().Get("WWW-Authenticate"); got != "Bearer" {
				t.Fatalf("WWW-Authenticate = %q, want Bearer", got)
			}
		})
	}
}

func TestRequireTokenPreflight(t *testing.T) {
	t.Parallel()
	handler := requireToken("secret", http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("preflight reached a route") }))
	for _, tt := range []struct {
		name, origin string
		wantStatus   int
	}{
		{name: "allowed", origin: "http://localhost:3000", wantStatus: http.StatusNoContent},
		{name: "untrusted", origin: "https://example.com", wantStatus: http.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/sessions", nil)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			req.Header.Set("Access-Control-Request-Headers", "Authorization")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusNoContent && w.Header().Get("Access-Control-Allow-Origin") != tt.origin {
				t.Fatal("missing CORS origin")
			}
		})
	}
}

func TestServeWaitsForHandlers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	entered, canceled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	defer unblock()
	srv := &http.Server{Handler: http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
		close(canceled)
		<-release
	})}
	defer srv.Close()
	ln, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- serve(ctx, srv, ln, time.Second)
	}()
	client := &http.Client{Timeout: 3 * time.Second}
	defer client.CloseIdleConnections()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+ln.Addr().String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		resp, reqErr := client.Do(req)
		if reqErr == nil {
			_ = resp.Body.Close()
		}
	}()
	t.Cleanup(func() {
		cancel()
		unblock()
		_ = srv.Close()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("serve cleanup: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("serve did not finish during cleanup")
		}
		select {
		case <-requestDone:
		case <-time.After(3 * time.Second):
			t.Error("client request did not finish during cleanup")
		}
	})
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not start")
	}
	cancel()
	select {
	case <-canceled:
	case <-time.After(3 * time.Second):
		t.Fatal("request was not canceled")
	}
	select {
	case err := <-done:
		t.Fatalf("returned before handler completed: %v", err)
	default:
	}
	unblock()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serve did not stop")
	}
	select {
	case <-requestDone:
	case <-time.After(3 * time.Second):
		t.Fatal("client request did not finish")
	}
}

func TestServeReportsListenerFailure(t *testing.T) {
	t.Parallel()
	ln, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	err = serve(t.Context(), &http.Server{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})}, ln, time.Second)
	if !errors.Is(err, net.ErrClosed) {
		t.Fatalf("error = %v", err)
	}
}

func TestServeForcesCloseAfterDeadline(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	entered, release, handlerDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	defer unblock()
	srv := &http.Server{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		defer close(handlerDone)
		close(entered)
		<-release
	})}
	defer srv.Close()
	ln, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- serve(ctx, srv, ln, 20*time.Millisecond)
	}()
	client := &http.Client{Timeout: 3 * time.Second}
	defer client.CloseIdleConnections()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+ln.Addr().String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		resp, reqErr := client.Do(req)
		if reqErr == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
	}()
	t.Cleanup(func() {
		cancel()
		unblock()
		_ = srv.Close()
		select {
		case err := <-done:
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("serve cleanup: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("serve did not finish during cleanup")
		}
		for _, ch := range []<-chan struct{}{requestDone, handlerDone} {
			select {
			case <-ch:
			case <-time.After(3 * time.Second):
				t.Error("request did not finish during cleanup")
			}
		}
	})
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("error = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown hung")
	}
	select {
	case <-requestDone:
	case <-time.After(3 * time.Second):
		t.Fatal("connection was not closed")
	}
	unblock()
	select {
	case <-handlerDone:
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not finish")
	}
}
