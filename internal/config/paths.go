package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var validName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func AIEnvsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ai-envs")
}

func EnvDir(name string) string {
	return filepath.Join(AIEnvsDir(), name)
}

func EnvYAML(name string) string {
	return filepath.Join(EnvDir(name), "ai-env.yaml")
}

func OpenCodeJSON(name string) string {
	return filepath.Join(EnvDir(name), "opencode.json")
}

func IsValidName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid name %q: must be lowercase alphanumeric with hyphens (e.g. backend-api)", name)
	}
	return nil
}
