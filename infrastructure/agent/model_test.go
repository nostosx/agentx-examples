package agent

import (
	"testing"

	"github.com/nostosx/agentx/model"
)

func TestModelFromEnv(t *testing.T) {
	for _, tt := range []struct {
		name         string
		env          map[string]string
		wantProvider string
		wantError    bool
	}{
		{name: "missing credentials", wantError: true},
		{name: "unknown provider", env: map[string]string{"AGENTX_PROVIDER": "unknown"}, wantError: true},
		{name: "explicit provider needs its own key", env: map[string]string{"AGENTX_PROVIDER": "openai", "ANTHROPIC_API_KEY": "fixture"}, wantError: true},
		{name: "provider priority", env: map[string]string{"ANTHROPIC_API_KEY": "fixture", "OPENAI_API_KEY": "fixture"}, wantProvider: "anthropic"},
		{name: "openai fallback", env: map[string]string{"OPENAI_API_KEY": "fixture"}, wantProvider: "openai"},
		{name: "google key alias", env: map[string]string{"GOOGLE_API_KEY": "fixture"}, wantProvider: "gcp.gemini"},
		{name: "explicit override", env: map[string]string{"AGENTX_PROVIDER": "gemini", "ANTHROPIC_API_KEY": "fixture", "GEMINI_API_KEY": "fixture"}, wantProvider: "gcp.gemini"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{"AGENTX_PROVIDER", "AGENTX_MODEL", "ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY", "GOOGLE_API_KEY", "ANTHROPIC_BASE_URL", "OPENAI_BASE_URL", "GEMINI_BASE_URL"} {
				t.Setenv(key, "")
			}
			for key, value := range tt.env {
				t.Setenv(key, value)
			}
			t.Setenv("AGENTX_MODEL", "fixture-model")
			m, err := ModelFromEnv()
			if tt.wantError {
				if err == nil || m != nil {
					t.Fatalf("model=%T err=%v", m, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			description := model.Describe(m)
			if description.Provider != tt.wantProvider || description.Model != "fixture-model" {
				t.Fatalf("description = %+v", description)
			}
		})
	}
}
