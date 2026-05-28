package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kapilratnani/aienv/internal/env"
)

func agentSkillPaths(agent string) []string {
	home, _ := os.UserHomeDir()
	switch agent {
	case "opencode":
		return []string{
			filepath.Join(home, ".config", "opencode", "skills"),
			filepath.Join(home, ".agents", "skills"),
			filepath.Join(home, ".claude", "skills"),
			filepath.Join(".config", "opencode", "skills"),
			filepath.Join(".agents", "skills"),
			filepath.Join(".claude", "skills"),
		}
	case "claude-code":
		return []string{
			filepath.Join(home, ".claude", "skills"),
			filepath.Join(".claude", "skills"),
		}
	default:
		return nil
	}
}

func SkillDir(name, agent string) string {
	return skillDir(name, agent)
}

func skillDir(name, agent string) string {
	for _, base := range agentSkillPaths(agent) {
		p := filepath.Join(base, name)
		if _, err := os.Stat(filepath.Join(p, "SKILL.md")); err == nil {
			return p
		}
	}
	return ""
}

func Verify(skill env.Skill, agent string) bool {
	return skillDir(skill.Name, agent) != ""
}

func VerifyAll(skills []env.Skill, agent string) []env.Skill {
	var missing []env.Skill
	for _, s := range skills {
		if !Verify(s, agent) {
			missing = append(missing, s)
		}
	}
	return missing
}

func Install(skill env.Skill, agent string) error {
	if skill.Source != "registry" {
		return fmt.Errorf("cannot auto-install skill %q: only registry source is supported", skill.Name)
	}
	if skill.Package == "" {
		return fmt.Errorf("cannot install skill %q: no package specified", skill.Name)
	}

	args := []string{"skills", "add", skill.Package, "--skill", skill.Name, "--agent", agent, "-y"}
	cmd := exec.Command("npx", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installing skill %q: %w", skill.Name, err)
	}
	return nil
}

func InstallAll(skills []env.Skill, agent string) error {
	for _, s := range skills {
		if err := Install(s, agent); err != nil {
			return err
		}
	}
	return nil
}
