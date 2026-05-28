package agents

import (
	"fmt"

	"github.com/kapilratnani/aienv/internal/env"
)

type AgentFile struct {
	Path    string
	Content []byte
}

type Agent interface {
	Name() string
	GenerateFiles(e *env.Env, cwd string) ([]AgentFile, error)
	ActivateCommand(envDir string, e *env.Env) string
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
