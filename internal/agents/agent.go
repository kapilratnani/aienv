package agents

import (
	"fmt"

	"github.com/kapilratnani/aienv/internal/env"
)

// AgentFile represents a file to be generated for an agent.
type AgentFile struct {
	Path    string
	Content []byte
}

// DockerConfig contains the configuration needed to run an agent in a Docker container.
type DockerConfig struct {
	Mounts     []string // "-v" arguments (config files, skills, etc.)
	EnvVars    []string // "-e" arguments
	Entrypoint []string // ["image:tag", "command", "--flag", ...]
}

// Agent represents an AI coding assistant that can be run in an environment.
type Agent interface {
	Name() string
	GenerateFiles(e *env.Env, cwd string) ([]AgentFile, error)
	DockerConfig(envDir string, e *env.Env, sessionID string) (*DockerConfig, error)
}

var registry = map[string]Agent{}

func Register(a Agent) {
	registry[a.Name()] = a
}

func Get(name string) (Agent, error) {
	a, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported agent %q", name)
	}
	return a, nil
}
