// Package http authenticates agent protocol requests.
package http

import (
	"context"
	"crypto/subtle"
	"net/http"
)

// All protocols, session operations and approval routes share this boundary.
func RequireToken(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Browser preflights carry no bearer token and must never reach a route.
		if r.Method == http.MethodOptions && r.Header.Get("Origin") == "http://localhost:3000" &&
			r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Last-Event-ID, A2A-Version, A2A-Extensions")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		values := r.Header.Values("Authorization")
		if token == "" || len(values) != 1 ||
			subtle.ConstantTimeCompare([]byte(values[0]), []byte("Bearer "+token)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CancelWithOwner cancels active model requests before application shutdown waits.
func CancelWithOwner(owner context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		stop := context.AfterFunc(owner, cancel)
		defer stop()
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
