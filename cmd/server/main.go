package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"

	application "github.com/nostosx/agentx-examples"
	"github.com/nostosx/agentx-examples/infrastructure/config"
	"github.com/nostosx/servex/app"
)

func main() {
	if err := run(); err != nil {
		slog.Error("assistant stopped", "error", err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := config.Parse("assistant", os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}
	ctx, stop := app.SignalContext(context.Background())
	defer stop()
	return application.Run(ctx, cfg)
}
