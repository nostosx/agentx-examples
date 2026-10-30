package command

import (
	"context"
	"fmt"

	"github.com/nostosx/servex/cqrs"
	"github.com/nostosx/servex/errorx"
)

type DeployBuildCommand struct {
	Env string `json:"env" jsonschema:"enum=staging,enum=production,description=Target environment"`
}
type DeployBuildResult struct{ Message string }
type DeployBuildHandler struct{}

func NewDeployBuildHandler() cqrs.Command[DeployBuildCommand, DeployBuildResult] {
	return &DeployBuildHandler{}
}

// Handle demonstrates a deployment response without performing an external write.
func (*DeployBuildHandler) Handle(ctx context.Context, in DeployBuildCommand) (DeployBuildResult, error) {
	if err := ctx.Err(); err != nil {
		return DeployBuildResult{}, err
	}
	if in.Env != "staging" && in.Env != "production" {
		return DeployBuildResult{}, errorx.New(errorx.InvalidArgument, "INVALID_ENVIRONMENT", "unknown deployment environment")
	}
	return DeployBuildResult{Message: fmt.Sprintf("Deployed the latest build to %s.", in.Env)}, nil
}
