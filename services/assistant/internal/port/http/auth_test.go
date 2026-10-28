package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
			handler := RequireToken("secret", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	handler := RequireToken("secret", http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("preflight reached a route") }))
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
