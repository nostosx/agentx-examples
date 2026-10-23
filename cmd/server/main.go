// Command server serves an agent over HTTP: the session API, the Vercel AI
// SDK useChat endpoint, AG-UI and A2A.
//
// Set AGENTX_SERVER_TOKEN to a random secret in the server and client environments.
//
//	go run ./cmd/server   # in agentx-examples
//	curl -N localhost:8080/sessions/demo/runs -H "Authorization: Bearer $AGENTX_SERVER_TOKEN" \
//	  -H 'Content-Type: application/json' \
//	  -d '{"input":"Deploy the latest build to staging"}'
//	curl localhost:8080/sessions/demo -H "Authorization: Bearer $AGENTX_SERVER_TOKEN"
//	curl -N localhost:8080/sessions/demo/resume -H "Authorization: Bearer $AGENTX_SERVER_TOKEN" \
//	  -H 'Content-Type: application/json' \
//	  -d '{"approvals":{"<call_id>":{"approved":true}}}'
package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nostosx/agentx-contrib/filestore"
	"github.com/nostosx/agentx-examples/internal/llm"
	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/server"
	"github.com/nostosx/agentx/tool"
)

type deployInput struct {
	Env string `json:"env" jsonschema:"enum=staging,enum=production,description=Target environment"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	if err := validateAddress(addr); err != nil {
		return err
	}
	token := os.Getenv("AGENTX_SERVER_TOKEN")
	if strings.TrimSpace(token) == "" || strings.ContainsAny(token, " \t\r\n") {
		return errors.New("set AGENTX_SERVER_TOKEN to a nonempty token without whitespace")
	}
	m, err := llm.New()
	if err != nil {
		return err
	}
	deploy := tool.Func("deploy", "Deploy the latest build.",
		func(_ context.Context, in deployInput) (string, error) {
			return fmt.Sprintf("Deployed the latest build to %s.", in.Env), nil
		}, tool.RequireApproval())

	cp, err := filestore.Checkpoints(".agentx/checkpoints")
	if err != nil {
		return err
	}
	store, err := filestore.Sessions(".agentx/sessions")
	if err != nil {
		return err
	}
	a := agent.MustNew("ops",
		agent.WithModel(m),
		agent.WithInstructions("You are a release assistant. Deploy builds when asked."),
		agent.WithTools(deploy),
		agent.WithCheckpointer(cp),
		agent.WithStreaming(),
	)
	srv := server.New(agent.NewRunner(a, store),
		server.WithCheckpointer(cp),
		server.WithCORS("http://localhost:3000"),
		server.WithA2A(server.A2AConfig{Name: "ops", Description: "Deploys builds after approval."}),
	)
	defer srv.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ln, err := new(net.ListenConfig).Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	hs := &http.Server{
		Handler:           requireToken(token, srv),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       time.Minute,
	}
	log.Printf("serving on http://%s", addr)
	return serve(ctx, hs, ln, 10*time.Second)
}

// This demo uses plaintext HTTP and local file state; keep it on loopback.
func validateAddress(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("listen address: %w", err)
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return errors.New("listen address must use localhost or a loopback IP")
	}
	return nil
}

// All protocols, session operations and approval routes share this boundary.
func requireToken(token string, next http.Handler) http.Handler {
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

func serve(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration) error {
	requests, cancelRequests := context.WithCancel(ctx)
	defer cancelRequests()
	srv.BaseContext = func(net.Listener) context.Context { return requests }
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ln) }()
	var serveErr error
	select {
	case <-ctx.Done():
		// Cancel model calls before waiting for HTTP handlers to finish.
		cancelRequests()
	case serveErr = <-done:
		cancelRequests()
		if !errors.Is(serveErr, http.ErrServerClosed) {
			return errors.Join(serveErr, srv.Close())
		}
		return nil
	}
	shutdown, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	shutdownErr := srv.Shutdown(shutdown)
	if shutdownErr != nil {
		shutdownErr = errors.Join(shutdownErr, srv.Close())
	}
	serveErr = <-done
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	return errors.Join(serveErr, shutdownErr)
}
