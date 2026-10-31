package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nostosx/agentx-examples/infrastructure/config"
	"github.com/nostosx/agentx-examples/services/knowledge"
)

func run() error {
	cfg, err := config.LoadPostgresRAG()
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(8)
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, 30*time.Second)
	defer cancel()
	result, err := knowledge.RunPostgresRAG(ctx, db, cfg)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "PostgreSQL RAG:", err)
		os.Exit(1)
	}
}
