package server

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nostosx/agentx-examples/infrastructure/config"
)

func testConfig() config.Config {
	return config.Config{Address: "127.0.0.1:0", Token: strings.Repeat("a", 32)}
}

func TestNewHonorsCancellationAndValidation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := New(ctx, testConfig(), nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := New(t.Context(), config.Config{}, nil); err == nil {
		t.Fatal("unsafe configuration accepted")
	}
	servers, err := New(t.Context(), testConfig(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if servers.HTTP.URL() != nil {
		t.Fatal("HTTP listener started during construction")
	}

}
