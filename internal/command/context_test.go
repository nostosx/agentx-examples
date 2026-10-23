package command

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestContextDeadline(t *testing.T) {
	ctx, cancel := Context(10 * time.Millisecond)
	defer cancel()
	select {
	case <-ctx.Done():
		if !errors.Is(context.Cause(ctx), context.DeadlineExceeded) {
			t.Fatalf("cause=%v", context.Cause(ctx))
		}
	case <-time.After(time.Second):
		t.Fatal("deadline did not cancel command")
	}
}

func TestContextCancel(t *testing.T) {
	for _, tt := range []struct {
		name    string
		timeout time.Duration
	}{
		{name: "signal only"},
		{name: "with deadline", timeout: time.Minute},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := Context(tt.timeout)
			cancel()
			if !errors.Is(context.Cause(ctx), context.Canceled) {
				t.Fatalf("cause=%v", context.Cause(ctx))
			}
		})
	}
}
