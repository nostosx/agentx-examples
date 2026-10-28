package config

import (
	"errors"
	"flag"
	"io"
	"strings"
	"testing"
)

func TestValidateRejectsUnsafeConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*Config)
	}{
		{"short token", func(c *Config) { c.Token = "short" }},
		{"whitespace token", func(c *Config) { c.Token = strings.Repeat(" ", 32) }},
		{"public HTTP", func(c *Config) { c.Address = "0.0.0.0:8080" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{Address: "127.0.0.1:0", Token: strings.Repeat("a", 32)}
			tc.edit(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("unsafe settings accepted")
			}
		})
	}
}

func TestParseRejectsArguments(t *testing.T) {
	for _, arg := range []string{"--bad", "unexpected", "--grpc-addr"} {
		t.Run(arg, func(t *testing.T) {
			if _, err := Parse("test", []string{arg}, io.Discard); err == nil {
				t.Fatal("invalid argument accepted")
			}
		})
	}
	if _, err := Parse("test", []string{"--help"}, io.Discard); !errors.Is(err, flag.ErrHelp) {
		t.Fatal(err)
	}
}

func TestParseDefaults(t *testing.T) {
	t.Setenv("API_TOKEN", strings.Repeat("a", 32))
	cfg, err := Parse("test", nil, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address != "127.0.0.1:8080" {
		t.Fatal(cfg)
	}

	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}
