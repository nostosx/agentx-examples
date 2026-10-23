// Command workflow runs agents in a graph: two drafts in parallel, a
// reviewer that picks one, and a conditional polishing step.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nostosx/agentx-examples/internal/command"
	"github.com/nostosx/agentx-examples/internal/llm"
	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/workflow"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := command.Context(5 * time.Minute)
	defer cancel()
	m, err := llm.New()
	if err != nil {
		return err
	}
	newAgent := func(name, instructions string) *agent.LLMAgent {
		return agent.MustNew(name, agent.WithModel(m), agent.WithInstructions(instructions))
	}
	formal := newAgent("formal", "Write formal, precise prose. One paragraph.")
	casual := newAgent("casual", "Write friendly, casual prose. One paragraph.")
	reviewer := newAgent("reviewer", "Compare two drafts. Reply with the better draft verbatim, then a final line 'VERDICT: OK' if it needs no edits or 'VERDICT: EDIT' otherwise.")
	editor := newAgent("editor", "Improve the draft's clarity and concision. Reply with the edited draft only.")

	g := workflow.NewGraph().
		AddNode("formal", workflow.AgentNode(formal, workflow.Template("Write about: {{topic}}"), "draft_a"),
			workflow.Retry(2, time.Second)).
		AddNode("casual", workflow.AgentNode(casual, workflow.Template("Write about: {{topic}}"), "draft_b"),
			workflow.Retry(2, time.Second)).
		AddNode("review", workflow.AgentNode(reviewer,
			workflow.Template("Draft A:\n{{draft_a}}\n\nDraft B:\n{{draft_b}}"), "review")).
		AddNode("edit", workflow.AgentNode(editor, workflow.Template("{{draft}}"), "final")).
		AddFunc("extract", func(_ context.Context, st *workflow.State) error {
			draft, verdict, err := parseReview(st.String("review"))
			if err != nil {
				return err
			}
			st.Set("draft", draft)
			st.Set("verdict", verdict)
			return nil
		}).
		AddFunc("accept", func(_ context.Context, st *workflow.State) error {
			st.Set("final", st.String("draft"))
			return nil
		}).
		// Both drafts run concurrently; review waits for both.
		AddEdge(workflow.Start, "formal").
		AddEdge(workflow.Start, "casual").
		AddEdge("formal", "review").
		AddEdge("casual", "review").
		AddChain("review", "extract").
		AddCondition("extract", func(_ context.Context, st *workflow.State) (string, error) {
			if st.String("verdict") == "EDIT" {
				return "edit", nil
			}
			return "ok", nil
		}, map[string]string{"edit": "edit", "ok": "accept"}).
		AddEdge("edit", workflow.End).
		AddEdge("accept", workflow.End)

	wf, err := g.Compile(workflow.OnEvent(func(ev workflow.Event) {
		if ev.Type == workflow.NodeEnd {
			log.Printf("%s done in %s", ev.Node, ev.Duration.Round(time.Millisecond))
		}
	}))
	if err != nil {
		return err
	}
	st, err := wf.Run(ctx, workflow.NewState(map[string]any{
		"topic": "why Go is a good fit for agent frameworks",
	}))
	if err != nil {
		return err
	}
	fmt.Println(st.String("final"))
	return nil
}

// parseReview requires a nonempty draft and one verdict on the final line.
func parseReview(text string) (string, string, error) {
	text = strings.TrimSpace(text)
	i := strings.LastIndexByte(text, '\n')
	if i < 0 {
		return "", "", errors.New("review: missing draft or final verdict line")
	}
	draft := strings.TrimSpace(text[:i])
	line := strings.TrimSpace(text[i+1:])
	if draft == "" || strings.Contains(draft, "VERDICT:") {
		return "", "", errors.New("review: empty draft or ambiguous verdict")
	}
	switch line {
	case "VERDICT: OK":
		return draft, "OK", nil
	case "VERDICT: EDIT":
		return draft, "EDIT", nil
	default:
		return "", "", errors.New("review: expected VERDICT: OK or VERDICT: EDIT on the final line")
	}
}
