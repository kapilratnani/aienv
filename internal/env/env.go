package env

import (
	"os"
	"path/filepath"
)

type Env struct {
	Name        string               `yaml:"name"`
	Agent       string               `yaml:"agent"`
	Model       string               `yaml:"model,omitempty"`
	Description string               `yaml:"description,omitempty"`
	Prompt      string               `yaml:"prompt,omitempty"`
	Workdir     string               `yaml:"workdir,omitempty"`
	MCPServers  map[string]MCPServer `yaml:"mcp"`
	Skills      []Skill              `yaml:"skills"`
	Rules       []Rule               `yaml:"rules"`
	Permissions *Permissions         `yaml:"permissions,omitempty"`
}

type MCPServer struct {
	Type    string            `yaml:"type"`
	Command []string          `yaml:"command,omitempty"`
	URL     string            `yaml:"url,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

type Skill struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Package string `yaml:"package,omitempty"`
	Path    string `yaml:"path,omitempty"`
}

type Rule struct {
	Path string `yaml:"path"`
}

type Permissions struct {
	Filesystem *FilesystemPermissions `yaml:"filesystem,omitempty"`
	Bash       map[string]string      `yaml:"bash,omitempty"`
	Network    *NetworkPermissions    `yaml:"network,omitempty"`
}

type FilesystemPermissions struct {
	Read map[string]string `yaml:"read,omitempty"`
	Edit map[string]string `yaml:"edit,omitempty"`
}

type NetworkPermissions struct {
	Allow []string `yaml:"allow,omitempty"`
	Deny  []string `yaml:"deny,omitempty"`
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
