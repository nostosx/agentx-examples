package application

import (
	"github.com/nostosx/agentx-examples/application/assistant/command"
	"github.com/nostosx/servex/cqrs"
)

type Command struct {
	DeployBuild cqrs.Command[command.DeployBuildCommand, command.DeployBuildResult]
}
type CommandDeps struct {
	DeployBuild cqrs.Command[command.DeployBuildCommand, command.DeployBuildResult]
	Middleware  []cqrs.CommandMiddleware[command.DeployBuildCommand, command.DeployBuildResult]
}

func NewCommand(deps CommandDeps) *Command {
	return &Command{DeployBuild: cqrs.ChainCommand("assistant.deploy_build", deps.DeployBuild, deps.Middleware...)}
}
