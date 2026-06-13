package env

import (
	"fmt"
	"os"

	"github.com/kapilratnani/aienv/internal/config"
	"gopkg.in/yaml.v3"
)

func Load(name string) (*Env, error) {
	yamlPath := config.EnvYAML(name)

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", yamlPath, err)
	}

	var e Env
	if err := yaml.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", yamlPath, err)
	}

	if e.Meta.Name == "" {
		e.Meta.Name = name
	}

	return &e, nil
}

func (e *Env) Validate() error {
	if e.Meta.Name == "" {
		return fmt.Errorf("env.name is required")
	}
	if len(e.Agent.Command) == 0 {
		return fmt.Errorf("agent.command is required")
	}
	if len(e.Agent.Mounts) == 0 {
		return fmt.Errorf("at least one mount is required (workdir)")
	}
	return nil
}

func (e *Env) Save() error {
	if err := e.Validate(); err != nil {
		return err
	}
	dir := config.EnvDir(e.Meta.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating env dir: %w", err)
	}

	data, err := yaml.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshalling env: %w", err)
	}

	yamlPath := config.EnvYAML(e.Meta.Name)
	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", yamlPath, err)
	}

	return nil
}
