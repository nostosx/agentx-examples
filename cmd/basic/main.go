// Command basic runs a single-turn agent.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nostosx/agentx-examples/internal/command"
	"github.com/nostosx/agentx-examples/internal/llm"
	"github.com/nostosx/agentx/agent"
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
	a := agent.MustNew("assistant",
		agent.WithModel(m),
		agent.WithInstructions("You are a concise assistant. Answer in at most three sentences."),
	)
	prompt := "What is an agent loop?"
	if len(os.Args) > 1 {
		prompt = strings.Join(os.Args[1:], " ")
	}
	answer, err := a.Ask(ctx, prompt)
	if err != nil {
		return err
	}
	fmt.Println(answer)
	return nil
}
