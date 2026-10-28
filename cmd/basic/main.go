// Command basic runs the basic agent example.
package main

import (
	"log"
	"os"
	"strings"
	"time"

	example "github.com/nostosx/agentx-examples/application/basic"
	agentinfra "github.com/nostosx/agentx-examples/infrastructure/agent"
	"github.com/nostosx/agentx-examples/infrastructure/command"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	ctx, cancel := command.Context(5 * time.Minute)
	defer cancel()
	prompt := "What is an agent loop?"
	if len(os.Args) > 1 {
		prompt = strings.Join(os.Args[1:], " ")
	}
	m, err := agentinfra.ModelFromEnv()
	if err != nil {
		return err
	}
	return example.Run(ctx, m, prompt)
}
