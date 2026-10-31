//go:build integration

package knowledge

import (
	"context"
	"crypto/rand"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nostosx/agentx-examples/infrastructure/config"
	pg "github.com/nostosx/servex-contrib/database/postgres"
)

func TestIntegrationPostgresRAG(t *testing.T) {
	dsn := os.Getenv("AGENTX_EXAMPLES_IT_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set AGENTX_EXAMPLES_IT_POSTGRES_DSN to an owned fixture")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := "ax_demo_" + strings.ToLower(rand.Text())[:16]
	ident, err := pg.Ident(schema)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := db.ExecContext(ctx, "DROP SCHEMA IF EXISTS "+ident+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	for range 2 {
		result, err := RunPostgresRAG(t.Context(), db, config.PostgresRAG{DSN: dsn, Schema: schema, Namespace: "demo", SearchEngine: "native"})
		if err != nil || result.VectorHits != 1 || result.TextHits != 1 || len(result.Citations) != 1 || strings.Contains(result.Content, "private/notes") {
			t.Fatalf("%+v: %v", result, err)
		}
	}
}
