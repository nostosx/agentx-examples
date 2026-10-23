// Command skills loads SKILL.md skills that the model can load on demand.
package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nostosx/agentx-examples/internal/command"
	"github.com/nostosx/agentx-examples/internal/llm"
	"github.com/nostosx/agentx/agent"
	"github.com/nostosx/agentx/loop"
	"github.com/nostosx/agentx/skill"
)

//go:embed skills
var skillFS embed.FS

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
	skills, err := loadSkills(*dir)
	if err != nil {
		return err
	}
	m, err := llm.New()
	if err != nil {
		return err
	}

	a := agent.MustNew("developer",
		agent.WithModel(m),
		agent.WithInstructions("You help with software development tasks."),
		agent.WithSkills(skills...),
	)
	res, err := loop.Collect(a.Run(ctx, agent.Text(
		"Write a commit message for: added retry with exponential backoff to the HTTP client.")),
		func(ev *loop.Event) {
			if ev.Type == loop.EventToolCall {
				fmt.Printf("[%s]\n", ev.ToolCall.Name)
			}
		})
	if err != nil {
		return err
	}
	fmt.Println(res.Output)
	return nil
}

// Local skills can register executable hooks, so loading them requires opt-in.
func loadSkills(dir string) ([]skill.Skill, error) {
	skills, err := skill.LoadFS(skillFS, "skills", "")
	if err != nil || dir == "" {
		return skills, err
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("local skills: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("local skills: %q is not a directory", dir)
	}
	local, err := skill.LoadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool, len(skills)+len(local))
	skills = append(skills, local...)
	for _, s := range skills {
		name := s.Metadata().Name
		if names[name] {
			return nil, fmt.Errorf("duplicate skill %q", name)
		}
		names[name] = true
	}
	return skills, nil
}
