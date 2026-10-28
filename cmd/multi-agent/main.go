// Command multi-agent runs the multi-agent agent example.
package main

import (
	"log"
	"time"

	example "github.com/nostosx/agentx-examples/application/multiagent"
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
	m, err := agentinfra.ModelFromEnv()
	if err != nil {
		return err
	}
	return example.Run(ctx, m)
}
