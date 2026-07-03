package env

import (
	"os"
	"path/filepath"
)

type Env struct {
	Meta        EnvMeta      `yaml:"env"`
	Agent       AgentConfig  `yaml:"agent"`
	Deps        DepsConfig   `yaml:"deps,omitempty"`
	Permissions *Permissions `yaml:"permissions,omitempty"`
	Audit       AuditConfig  `yaml:"audit,omitempty"`
}

type EnvMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type AgentConfig struct {
	Install        []string          `yaml:"install"`
	Command        []string          `yaml:"command"`
	Args           []string          `yaml:"args,omitempty"`
	PromptFlag     string            `yaml:"prompt_flag,omitempty"`
	ExitSubcommand string            `yaml:"exit_subcommand,omitempty"`
	Env            map[string]string `yaml:"env,omitempty"`
	Mounts         []Mount           `yaml:"mounts"`
}

type Mount struct {
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	Writable bool   `yaml:"writable,omitempty"`
}

type DepsConfig struct {
	Packages []string `yaml:"packages,omitempty"`
	Custom   []string `yaml:"custom,omitempty"`
}

type Permissions struct {
	Network    *NetworkPermissions `yaml:"network,omitempty"`
	Filesystem []string            `yaml:"filesystem,omitempty"`
}

type NetworkPermissions struct {
	Allow []string `yaml:"allow,omitempty"`
	Deny  []string `yaml:"deny,omitempty"`
	Learn bool     `yaml:"learn,omitempty"`
}

type AuditConfig struct {
	Persist bool     `yaml:"persist"`
	Capture []string `yaml:"capture,omitempty"`
}

func ExpandTilde(path string) string {
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if len(path) > 1 && path[0] == '~' && path[1] == '/' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func (e *Env) ApplyPrompt(prompt string) {
	if prompt == "" {
		return
	}
	if e.Agent.PromptFlag != "" {
		e.Agent.Args = append(e.Agent.Args, e.Agent.PromptFlag, prompt)
	} else {
		e.Agent.Args = append(e.Agent.Args, prompt)
	}
}

func (m *Mount) ResolveSource() string {
	return ExpandTilde(m.Source)
}

func (m *Mount) ResolveTarget() string {
	if len(m.Target) > 0 && m.Target[0] == '~' {
		return "/home/agent" + m.Target[1:]
	}
	return m.Target
}
