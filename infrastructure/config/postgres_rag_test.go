package config

import "testing"

func TestPostgresRAGConfiguration(t *testing.T) {
	t.Setenv("AGENTX_DEMO_POSTGRES_DSN", "postgres://fixture")
	for _, schema := range []string{"", "public", "ax_demo_bad-name", "ax_demo_abcdefghijklmnopqrstuvwxyz"} {
		t.Setenv("AGENTX_DEMO_POSTGRES_SCHEMA", schema)
		if _, err := LoadPostgresRAG(); err == nil {
			t.Fatalf("accepted schema %q", schema)
		}
	}
	t.Setenv("AGENTX_DEMO_POSTGRES_SCHEMA", "ax_demo_test")
	c, err := LoadPostgresRAG()
	if err != nil || c.SearchEngine != "native" || c.Namespace != "demo" {
		t.Fatalf("%+v: %v", c, err)
	}
}
