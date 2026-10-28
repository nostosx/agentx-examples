// Command streaming runs the streaming agent example.
package main

import (
	"log"
	"os"

	example "github.com/nostosx/agentx-examples/application/streaming"
	agentinfra "github.com/nostosx/agentx-examples/infrastructure/agent"
	"github.com/nostosx/agentx-examples/infrastructure/command"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	ctx, cancel := command.Context(0)
	defer cancel()
	m, err := agentinfra.ModelFromEnv()
	if err != nil {
		return err
	}
	return example.Run(ctx, m, os.Stdin, os.Stdout)
}
