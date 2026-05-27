package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kapilratnani/aienv/internal/env"
)

func skillDir(name string) string {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".config", "opencode", "skills", name),
		filepath.Join(home, ".claude", "skills", name),
		filepath.Join(home, ".claude", "skills", name),
		filepath.Join(".agents", "skills", name),
		filepath.Join(home, ".agents", "skills", name),
		filepath.Join(".claude", "skills", name),
	}
	for _, p := range paths {
		if _, err := os.Stat(filepath.Join(p, "SKILL.md")); err == nil {
			return p
		}
	}
	return ""
}

func Verify(skill env.Skill) bool {
	return skillDir(skill.Name) != ""
}

func VerifyAll(skills []env.Skill) []env.Skill {
	var missing []env.Skill
	for _, s := range skills {
		if !Verify(s) {
			missing = append(missing, s)
		}
	}
	return missing
}

func Install(skill env.Skill) error {
	if skill.Source != "registry" {
		return fmt.Errorf("cannot auto-install skill %q: only registry source is supported", skill.Name)
	}
	if skill.Package == "" {
		return fmt.Errorf("cannot install skill %q: no package specified", skill.Name)
	}

	args := []string{"skills", "add", skill.Package, "--skill", skill.Name, "-y"}
	cmd := exec.Command("npx", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installing skill %q: %w", skill.Name, err)
	}
	return nil
}

func InstallAll(skills []env.Skill) error {
	for _, s := range skills {
		if err := Install(s); err != nil {
			return err
		}
	}
	return nil
}
