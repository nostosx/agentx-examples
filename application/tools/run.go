// Package tools demonstrates agentx tools behavior.
package tools

import (
	"context"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"
	"sync"

	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/hook"
	"github.com/nostosx/agentx/loop"
	"github.com/nostosx/agentx/model"
	"github.com/nostosx/agentx/tool"
)

type weatherArgs struct {
	City string `json:"city" jsonschema:"City name, e.g. Tokyo"`
}

type calcArgs struct {
	Op string  `json:"op" jsonschema:"enum=add,enum=sub,enum=mul,enum=div,enum=pow"`
	A  float64 `json:"a"`
	B  float64 `json:"b"`
}

type noteArgs struct {
	Text string `json:"text" jsonschema:"Note to save"`
}

func Run(ctx context.Context, m model.Model) error {
	weather := tool.Func("get_weather", "Get the current weather for a city.",
		func(_ context.Context, in weatherArgs) (string, error) {
			// A real tool would call a weather API.
			return fmt.Sprintf("%s: 22°C, partly cloudy", in.City), nil
		}, tool.ReadOnly())

	calc := tool.Func("calculate", "Evaluate a binary arithmetic operation.",
		calculate, tool.ReadOnly())

	var notes noteStore
	save := tool.Func("save_note", "Save a note for the user.",
		notes.save, tool.RequireApproval())

	a := agent.MustNew("tool_user",
		agent.WithModel(m),
		agent.WithInstructions("Use the tools to answer. Save a short note summarizing your answer."),
		agent.WithTools(weather, calc, save),
		// Log every tool call before it runs.
		agent.WithHook(hook.PreToolUse, "*", hook.Func(func(_ context.Context, ev *hook.Event) (*hook.Result, error) {
			log.Printf("hook: %s", ev.ToolName())
			return nil, nil
		})),
		// This demo auto-approves calls; applications should apply their own policy.
		agent.WithApprover(loop.ApproverFunc(func(_ context.Context, req *loop.ApprovalRequest) (loop.Approval, error) {
			log.Printf("approving %s", req.Call.Name)
			return loop.Approve, nil
		})),
	)

	res, err := loop.Collect(a.Run(ctx, agent.Text(
		"What's the weather in Tokyo, and what is 2 to the power of 10?")),
		func(ev *loop.Event) {
			if ev.Type == loop.EventToolResult {
				fmt.Printf("  -> %s: %s\n", ev.ToolResult.Name, ev.ToolResult.Content)
			}
		})
	if err != nil {
		return err
	}
	fmt.Println(res.Output)
	fmt.Printf("notes: %s\n", strings.Join(notes.snapshot(), "; "))
	fmt.Printf("turns: %d, tokens: %d in / %d out\n", res.Turns, res.Usage.InputTokens, res.Usage.OutputTokens)
	return nil
}

func calculate(_ context.Context, in calcArgs) (*tool.Result, error) {
	var value float64
	switch in.Op {
	case "add":
		value = in.A + in.B
	case "sub":
		value = in.A - in.B
	case "mul":
		value = in.A * in.B
	case "div":
		if in.B == 0 {
			return tool.Errorf("division by zero"), nil
		}
		value = in.A / in.B
	case "pow":
		value = math.Pow(in.A, in.B)
	default:
		return tool.Errorf("unknown op %q", in.Op), nil
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return tool.Errorf("result is not finite"), nil
	}
	return tool.Text(fmt.Sprint(value)), nil
}

// noteStore supports concurrent callbacks and returns independent snapshots.
type noteStore struct {
	mu    sync.Mutex
	notes []string
}

func (s *noteStore) save(_ context.Context, in noteArgs) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notes = append(s.notes, in.Text)
	return "saved", nil
}

func (s *noteStore) snapshot() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.notes)
}
