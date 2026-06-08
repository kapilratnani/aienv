package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kapilratnani/aienv/internal/env"
)

func TestAgentSkillPaths(t *testing.T) {
	t.Run("opencode paths", func(t *testing.T) {
		paths := agentSkillPaths("opencode")
		if len(paths) == 0 {
			t.Fatal("expected non-empty paths for opencode")
		}
		foundHome := false
		for _, p := range paths {
			if strings.Contains(p, ".config/opencode/skills") && filepath.IsAbs(p) {
				foundHome = true
			}
		}
		if !foundHome {
			t.Error("expected absolute .config/opencode/skills path")
		}
	})

	t.Run("claude-code paths", func(t *testing.T) {
		paths := agentSkillPaths("claude-code")
		if len(paths) == 0 {
			t.Fatal("expected non-empty paths for claude-code")
		}
		hasClaudeSkills := false
		for _, p := range paths {
			if strings.Contains(p, ".claude/skills") {
				hasClaudeSkills = true
			}
		}
		if !hasClaudeSkills {
			t.Error("expected claude-code paths to include .claude/skills")
		}
	})

	t.Run("unknown agent", func(t *testing.T) {
		paths := agentSkillPaths("unknown")
		if paths != nil {
			t.Error("expected nil paths for unknown agent")
		}
	})
}

func TestVerify(t *testing.T) {
	tmpDir := t.TempDir()
	setHome(t, tmpDir)

	skillDir := filepath.Join(tmpDir, ".config", "opencode", "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("skill found", func(t *testing.T) {
		got := Verify(env.Skill{Name: "test-skill"}, "opencode")
		if !got {
			t.Error("expected skill to be found")
		}
	})

	t.Run("skill not found", func(t *testing.T) {
		got := Verify(env.Skill{Name: "nonexistent"}, "opencode")
		if got {
			t.Error("expected skill not to be found")
		}
	})
}

func TestVerifyAll(t *testing.T) {
	tmpDir := t.TempDir()
	setHome(t, tmpDir)

	skillDir := filepath.Join(tmpDir, ".config", "opencode", "skills", "installed-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	skills := []env.Skill{
		{Name: "installed-skill"},
		{Name: "missing-skill"},
	}

	missing := VerifyAll(skills, "opencode")
	if len(missing) != 1 {
		t.Fatalf("expected 1 missing skill, got %d", len(missing))
	}
	if missing[0].Name != "missing-skill" {
		t.Errorf("expected missing skill 'missing-skill', got %q", missing[0].Name)
	}
}

func TestInstallValidation(t *testing.T) {
	t.Run("non-registry source", func(t *testing.T) {
		err := Install(env.Skill{Name: "test", Source: "file"}, "opencode")
		if err == nil {
			t.Fatal("expected error for non-registry source")
		}
	})

	t.Run("empty package", func(t *testing.T) {
		err := Install(env.Skill{Name: "test", Source: "registry", Package: ""}, "opencode")
		if err == nil {
			t.Fatal("expected error for empty package")
		}
	})
}

func setHome(t *testing.T, dir string) {
	t.Helper()
	orig := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", orig) })
}
