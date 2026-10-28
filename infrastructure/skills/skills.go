// Package skills loads embedded and explicitly trusted local skills.
package skills

import (
	"embed"
	"fmt"
	"os"

	"github.com/nostosx/agentx/skill"
)

//go:embed skills
var skillFS embed.FS

// Local skills can register executable hooks, so loading them requires opt-in.
func Load(dir string) ([]skill.Skill, error) {
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
