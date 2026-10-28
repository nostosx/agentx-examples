// Package command provides the cancellation boundary for example commands.
package command

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Context cancels on SIGINT or SIGTERM and, when positive, after timeout.
// The caller must call the returned cancel function to release signal and timer resources.
func Context(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	if timeout <= 0 {
		return ctx, stop
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	return ctx, func() {
		cancel()
		stop()
	}
}
