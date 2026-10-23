package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadSkillsEmbeddedOnly(t *testing.T) {
	t.Parallel()
	skills, err := loadSkills("")
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 || skills[0].Metadata().Name != "commit-message" {
		t.Fatalf("skills=%v", skills)
	}
}

func TestLoadSkillsInvalidDirectory(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ name, path string }{
		{name: "missing", path: filepath.Join(root, "missing")},
		{name: "file", path: file},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := loadSkills(tt.path); err == nil {
				t.Fatal("invalid directory accepted")
			}
		})
	}
}

func TestLoadSkillsExplicitDirectory(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, skillName, document string
		wantError                 bool
	}{
		{name: "local skill", skillName: "local", document: "---\nname: local\ndescription: Local fixture\n---\nFixture."},
		{name: "duplicate embedded name", skillName: "commit-message", document: "---\nname: commit-message\ndescription: Duplicate\n---\nFixture.", wantError: true},
		{name: "malformed skill", skillName: "bad", document: "missing frontmatter", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			dir := filepath.Join(root, tt.skillName)
			if err := os.Mkdir(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(tt.document), 0o600); err != nil {
				t.Fatal(err)
			}
			skills, err := loadSkills(root)
			if tt.wantError {
				if err == nil {
					t.Fatal("invalid skill accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, s := range skills {
				names = append(names, s.Metadata().Name)
			}
			slices.Sort(names)
			want := []string{"commit-message", "local"}
			if !slices.Equal(names, want) {
				t.Fatalf("skill names = %q, want %q", names, want)
			}
		})
	}
}
