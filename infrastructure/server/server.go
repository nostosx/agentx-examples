// Package server constructs the process transports without starting them.
package server

import (
	"context"
	"time"

	"github.com/nostosx/servex/health"
	"github.com/nostosx/servex/transport/httpx"

	"github.com/nostosx/agentx-examples/infrastructure/config"
)

// Servers contains only the listeners enabled for this project.
type Servers struct {
	HTTP *httpx.Server
}

// New validates configuration and constructs transports. The caller mounts
// handlers and passes the transports to app.WithHooks in explicit order.
func New(ctx context.Context, cfg config.Config, check health.Checker) (*Servers, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	http := httpx.NewServer(httpx.WithAddress(cfg.Address), httpx.WithHealth(check),
		httpx.WithReadHeaderTimeout(5*time.Second), httpx.WithReadTimeout(30*time.Second),
		httpx.WithWriteTimeout(0), httpx.WithIdleTimeout(time.Minute), httpx.WithMaxHeaderBytes(16<<10), httpx.WithMaxBodyBytes(64<<10))
	return &Servers{HTTP: http}, nil
}
