// Package assistant assembles the release assistant and its approval tool.
package assistant

import (
	"context"
	"fmt"

	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/loop"
	"github.com/nostosx/agentx/model"
	"github.com/nostosx/agentx/session"
	"github.com/nostosx/agentx/tool"
)

type deployInput struct {
	Env string `json:"env" jsonschema:"enum=staging,enum=production,description=Target environment"`
}

func New(m model.Model, checkpoints loop.Checkpointer, sessions session.Store) *agent.Runner {
	deploy := tool.Func("deploy", "Deploy the latest build.", func(_ context.Context, in deployInput) (string, error) {
		return fmt.Sprintf("Deployed the latest build to %s.", in.Env), nil
	}, tool.RequireApproval())
	a := agent.MustNew("ops", agent.WithModel(m), agent.WithInstructions("You are a release assistant. Deploy builds when asked."), agent.WithTools(deploy), agent.WithCheckpointer(checkpoints), agent.WithStreaming())
	return agent.NewRunner(a, sessions)
}
