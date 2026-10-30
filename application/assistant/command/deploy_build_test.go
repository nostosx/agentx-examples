package command

import (
	"context"
	"errors"
	"testing"

	"github.com/nostosx/servex/errorx"
)

func TestDeployBuild(t *testing.T) {
	t.Parallel()
	handler := &DeployBuildHandler{}
	for _, env := range []string{"staging", "production"} {
		result, err := handler.Handle(t.Context(), DeployBuildCommand{Env: env})
		if err != nil || result.Message != "Deployed the latest build to "+env+"." {
			t.Fatalf("%+v %v", result, err)
		}
	}
	if _, err := handler.Handle(t.Context(), DeployBuildCommand{Env: "unknown"}); errorx.CodeOf(err) != errorx.InvalidArgument {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := handler.Handle(ctx, DeployBuildCommand{Env: "staging"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
