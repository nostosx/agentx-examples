// Package skills demonstrates agentx skills behavior.
package skills

import (
	"context"
	"fmt"

	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/loop"
	"github.com/nostosx/agentx/model"
	"github.com/nostosx/agentx/skill"
)

func Run(ctx context.Context, m model.Model, skills []skill.Skill) error {

	a := agent.MustNew("developer",
		agent.WithModel(m),
		agent.WithInstructions("You help with software development tasks."),
		agent.WithSkills(skills...),
	)
	res, err := loop.Collect(a.Run(ctx, agent.Text(
		"Write a commit message for: added retry with exponential backoff to the HTTP client.")),
		func(ev *loop.Event) {
			if ev.Type == loop.EventToolCall {
				fmt.Printf("[%s]\n", ev.ToolCall.Name)
			}
		})
	if err != nil {
		return err
	}
	fmt.Println(res.Output)
	return nil
}
