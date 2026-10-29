// Package application assembles the monolith host.
package application

import (
	"context"

	"github.com/nostosx/agentx-examples/infrastructure/config"
	"github.com/nostosx/agentx-examples/services/assistant"
)

// Run hosts the assistant service using its shared assembly and lifecycle.
func Run(ctx context.Context, cfg config.Config) error {
	return assistant.Run(ctx, cfg)
}
