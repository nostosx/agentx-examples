// Command skills runs the skills agent example.
package main

import (
	"flag"
	"log"
	"time"

	example "github.com/nostosx/agentx-examples/application/skills"
	agentinfra "github.com/nostosx/agentx-examples/infrastructure/agent"
	"github.com/nostosx/agentx-examples/infrastructure/command"
	"github.com/nostosx/agentx-examples/infrastructure/skills"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	ctx, cancel := command.Context(5 * time.Minute)
	defer cancel()
	dir := flag.String("skills-dir", "", "optional directory of trusted local skills")
	flag.Parse()
	loaded, err := skills.Load(*dir)
	if err != nil {
		return err
	}
	m, err := agentinfra.ModelFromEnv()
	if err != nil {
		return err
	}
	return example.Run(ctx, m, loaded)
}
