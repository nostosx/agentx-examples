// Package multiagent demonstrates agentx multi-agent behavior.
package multiagent

import (
	"context"
	"fmt"
	"log"

	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/hook"
	"github.com/nostosx/agentx/loop"
	"github.com/nostosx/agentx/model"
)

func Run(ctx context.Context, m model.Model) error {

	researcher := agent.MustNew("researcher",
		agent.WithModel(m),
		agent.WithDescription("Gathers facts about a topic and reports them as a bullet list."),
		agent.WithInstructions("You research topics. Reply with 3-5 factual bullet points."),
	)
	writer := agent.MustNew("writer",
		agent.WithModel(m),
		agent.WithDescription("Turns notes into a polished short paragraph."),
		agent.WithInstructions("You are a technical writer. Write one clear paragraph from the notes you receive."),
	)

	coordinator := agent.MustNew("coordinator",
		agent.WithModel(m),
		agent.WithInstructions("Break the task down. Use researcher for facts, then writer for the final text. Return the writer's paragraph."),
		agent.WithSubAgents(researcher, writer),
		agent.WithHook(hook.SubagentStop, "*", hook.Func(func(_ context.Context, ev *hook.Event) (*hook.Result, error) {
			log.Printf("sub-agent %s finished (%d chars)", ev.Agent, len(ev.Message))
			return nil, nil
		})),
	)

	res, err := loop.Collect(coordinator.Run(ctx,
		agent.Text("Explain what the Model Context Protocol is.")),
		func(ev *loop.Event) {
			if ev.Type == loop.EventToolCall {
				fmt.Printf("-> delegating to %s\n", ev.ToolCall.Name)
			}
		})
	if err != nil {
		return err
	}
	fmt.Println(res.Output)
	return nil
}
