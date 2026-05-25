package env

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(name string) (*Env, error) {
	envDir := filepath.Join(os.Getenv("HOME"), ".ai-envs", name)
	yamlPath := filepath.Join(envDir, "ai-env.yaml")

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", yamlPath, err)
	}

	var e Env
	if err := yaml.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", yamlPath, err)
	}

	e.Name = name
	return &e, nil
}

func (e *Env) Save() error {
	envDir := filepath.Join(os.Getenv("HOME"), ".ai-envs", e.Name)
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return fmt.Errorf("creating env dir: %w", err)
	}

	data, err := yaml.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshalling env: %w", err)
	}

	yamlPath := filepath.Join(envDir, "ai-env.yaml")
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", yamlPath, err)
	}

	return nil
}

func (e *Env) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("name is required")
	}
	if e.Agent == "" {
		return fmt.Errorf("agent is required")
	}
	if e.Agent != "opencode" {
		return fmt.Errorf("unsupported agent %q (only opencode is supported)", e.Agent)
	}
	return nil
}
