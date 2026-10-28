// Package basic demonstrates agentx basic behavior.
package basic

import (
	"context"
	"fmt"

	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/model"
)

func Run(ctx context.Context, m model.Model, prompt string) error {
	a := agent.MustNew("assistant",
		agent.WithModel(m),
		agent.WithInstructions("You are a concise assistant. Answer in at most three sentences."),
	)
	answer, err := a.Ask(ctx, prompt)
	if err != nil {
		return err
	}
	fmt.Println(answer)
	return nil
}
