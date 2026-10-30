// Package assistant assembles the release assistant and its approval tool.
package assistant

import (
	"context"

	"github.com/nostosx/agentx-examples/application/assistant/command"
	"github.com/nostosx/servex/cqrs"

	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/loop"
	"github.com/nostosx/agentx/model"
	"github.com/nostosx/agentx/session"
	"github.com/nostosx/agentx/tool"
)

func New(m model.Model, checkpoints loop.Checkpointer, sessions session.Store, deployBuild cqrs.Command[command.DeployBuildCommand, command.DeployBuildResult]) *agent.Runner {
	deploy := tool.Func("deploy", "Deploy the latest build.", func(ctx context.Context, in command.DeployBuildCommand) (string, error) {
		result, err := deployBuild.Handle(ctx, in)
		return result.Message, err
	}, tool.RequireApproval())
	a := agent.MustNew("ops", agent.WithModel(m), agent.WithInstructions("You are a release assistant. Deploy builds when asked."), agent.WithTools(deploy), agent.WithCheckpointer(checkpoints), agent.WithStreaming())
	return agent.NewRunner(a, sessions)
}
