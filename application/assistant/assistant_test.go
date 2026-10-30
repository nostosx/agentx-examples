package assistant

import (
	"context"
	"testing"

	"github.com/nostosx/agentx-examples/application/assistant/command"
	"github.com/nostosx/agentx/loop"
	"github.com/nostosx/agentx/model/modeltest"
	"github.com/nostosx/agentx/session"
	"github.com/nostosx/servex/cqrs"
)

func TestApprovalInvokesInjectedCommand(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		approval loop.Approval
		want     int
	}{{"approve", loop.Approve, 1}, {"deny", loop.Deny, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			deploy := cqrs.ChainCommand("assistant.deploy_build", cqrs.CommandFunc[command.DeployBuildCommand, command.DeployBuildResult](func(_ context.Context, input command.DeployBuildCommand) (command.DeployBuildResult, error) {
				calls++
				if input.Env != "staging" {
					t.Fatal(input)
				}
				return command.DeployBuildResult{Message: "fixture-deployed"}, nil
			}))
			m := modeltest.New(modeltest.CallTools(modeltest.Call{ID: "deploy-1", Name: "deploy", Args: map[string]string{"env": "staging"}}), modeltest.Reply("done"))
			runner := New(m, loop.NewMemoryCheckpointer(), session.NewMemoryStore(), deploy)
			result, err := runner.RunText(t.Context(), "approval-demo", "deploy staging")
			if err != nil || result.Status != loop.StatusInterrupted || calls != 0 {
				t.Fatalf("before approval: %v %v calls=%d", result, err, calls)
			}
			result, err = loop.Collect(runner.Resume(t.Context(), "approval-demo", map[string]loop.Approval{"deploy-1": tc.approval}), nil)
			if err != nil || result.Output != "done" {
				t.Fatalf("resume: %v %v", result, err)
			}
			if calls != tc.want {
				t.Fatalf("command calls=%d want=%d", calls, tc.want)
			}

		})
	}
}
