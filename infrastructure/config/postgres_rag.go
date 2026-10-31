package config

import (
	"errors"
	"os"
	"regexp"
)

// PostgresRAG is explicit configuration for an owned PostgreSQL demo schema.
type PostgresRAG struct{ DSN, Schema, Namespace, SearchEngine string }

func LoadPostgresRAG() (PostgresRAG, error) {
	c := PostgresRAG{DSN: os.Getenv("AGENTX_DEMO_POSTGRES_DSN"), Schema: os.Getenv("AGENTX_DEMO_POSTGRES_SCHEMA"), Namespace: "demo", SearchEngine: os.Getenv("AGENTX_DEMO_SEARCH_ENGINE")}
	if c.SearchEngine == "" {
		c.SearchEngine = "native"
	}
	return c, c.Validate()
}

// Validate confines the demonstration to its explicitly owned schema.
func (c PostgresRAG) Validate() error {
	if c.DSN == "" || !regexp.MustCompile(`^ax_demo_[a-z0-9_]{1,20}$`).MatchString(c.Schema) {
		return errors.New("set AGENTX_DEMO_POSTGRES_DSN and a dedicated AGENTX_DEMO_POSTGRES_SCHEMA (ax_demo_ prefix)")
	}
	if c.Namespace == "" {
		return errors.New("demo namespace is required")
	}
	return nil
}
