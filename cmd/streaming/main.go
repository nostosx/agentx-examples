// Command streaming prints model output as it is generated and keeps a
// multi-turn conversation in a session.
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nostosx/agentx-examples/internal/command"
	"github.com/nostosx/agentx-examples/internal/llm"
	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/loop"
	"github.com/nostosx/agentx/model"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	m, err := llm.New()
	if err != nil {
		return err
	}
	a := agent.MustNew("chat",
		agent.WithModel(m),
		agent.WithInstructions("You are a friendly assistant."),
		agent.WithStreaming(),
	)
	ctx, cancel := command.Context(0)
	defer cancel()
	return chat(ctx, agent.NewRunner(a, nil), os.Stdin, os.Stdout)
}

func chat(ctx context.Context, runner *agent.Runner, input io.Reader, output io.Writer) error {
	const sessionID = "demo"

	if err := context.Cause(ctx); err != nil {
		return err
	}
	if _, err := fmt.Fprint(output, "> "); err != nil {
		return err
	}
	ctx, stopInput := context.WithCancel(ctx)
	defer stopInput()
	type inputLine struct {
		text string
		err  error
	}
	lines := make(chan inputLine)
	// Some OS stdin reads cannot be interrupted by Close. The command owns
	// one reader goroutine for its process lifetime; chat can still cancel.
	go func() {
		defer close(lines)
		in := bufio.NewScanner(input)
		for in.Scan() {
			select {
			case lines <- inputLine{text: in.Text()}:
			case <-ctx.Done():
				return
			}
		}
		if err := in.Err(); err != nil {
			select {
			case lines <- inputLine{err: err}:
			case <-ctx.Done():
			}
		}
	}()
	for {
		var next inputLine
		var ok bool
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case next, ok = <-lines:
			if err := context.Cause(ctx); err != nil {
				return err
			}
			if !ok || next.err != nil {
				return next.err
			}
		}
		line := strings.TrimSpace(next.text)
		if line == "" {
			if _, err := fmt.Fprint(output, "> "); err != nil {
				return err
			}
			continue
		}
		if line == "/exit" {
			return nil
		}
		turn, cancel := context.WithTimeout(ctx, 5*time.Minute)
		for ev, err := range runner.Run(turn, sessionID, model.UserText(line)) {
			if err != nil {
				cancel()
				return err
			}
			switch ev.Type {
			case loop.EventDelta:
				_, err = fmt.Fprint(output, ev.Text())
			case loop.EventRunEnd:
				_, err = fmt.Fprintf(output, "\n[%s, %d output tokens]\n", ev.Result.Status, ev.Result.Usage.OutputTokens)
			}
			if err != nil {
				cancel()
				return err
			}
		}
		cancel()
		if _, err := fmt.Fprint(output, "> "); err != nil {
			return err
		}
	}
}
