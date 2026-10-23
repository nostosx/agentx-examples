// Package llm selects a model provider for the examples from the
// environment.
//
// AGENTX_PROVIDER chooses anthropic, openai or gemini; otherwise the first
// provider with an API key set is used. AGENTX_MODEL overrides the model.
package llm

import (
	"errors"
	"fmt"
	"os"

	"github.com/nostosx/agentx-contrib/provider/anthropic"
	"github.com/nostosx/agentx-contrib/provider/gemini"
	"github.com/nostosx/agentx-contrib/provider/openai"
	"github.com/nostosx/agentx/model"
)

// New constructs the provider selected by the environment without making a request.
// It returns a nil model when provider configuration fails.
func New() (model.Model, error) {
	provider := os.Getenv("AGENTX_PROVIDER")
	if provider == "" {
		switch {
		case os.Getenv("ANTHROPIC_API_KEY") != "":
			provider = "anthropic"
		case os.Getenv("OPENAI_API_KEY") != "":
			provider = "openai"
		case os.Getenv("GEMINI_API_KEY") != "" || os.Getenv("GOOGLE_API_KEY") != "":
			provider = "gemini"
		default:
			return nil, errors.New("set ANTHROPIC_API_KEY, OPENAI_API_KEY or GEMINI_API_KEY")
		}
	}
	name := os.Getenv("AGENTX_MODEL")
	var m model.Model
	var err error
	switch provider {
	case "anthropic":
		var opts []anthropic.Option
		if name != "" {
			opts = append(opts, anthropic.WithModel(name))
		}
		m, err = anthropic.New(opts...)
	case "openai":
		var opts []openai.Option
		if name != "" {
			opts = append(opts, openai.WithModel(name))
		}
		m, err = openai.New(opts...)
	case "gemini":
		var opts []gemini.Option
		if name != "" {
			opts = append(opts, gemini.WithModel(name))
		}
		m, err = gemini.New(opts...)
	default:
		return nil, fmt.Errorf("unknown AGENTX_PROVIDER %q", provider)
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}
