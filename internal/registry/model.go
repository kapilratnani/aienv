package registry

import "context"

type MCPServerItem struct {
	Name         string       `json:"name"`
	DisplayName  string       `json:"displayName,omitempty"`
	Command      string       `json:"command,omitempty"`
	Type         string       `json:"type"`
	URL          string       `json:"url,omitempty"`
	Description  string       `json:"description,omitempty"`
	Package      string       `json:"package,omitempty"`
	RegistryType string       `json:"registryType,omitempty"`
	EnvVars      []EnvVarItem `json:"env,omitempty"`
	Source       string       `json:"-"`
}

type EnvVarItem struct {
	Key         string `json:"key"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type SkillItem struct {
	Name        string `json:"name"`
	Package     string `json:"package"`
	Description string `json:"description,omitempty"`
	Installs    int    `json:"installs,omitempty"`
	Source      string `json:"-"`
}

type MCPRegistry interface {
	Search(ctx context.Context, query string, limit int) ([]MCPServerItem, error)
}

type SkillRegistry interface {
	Search(ctx context.Context, query string, limit int) ([]SkillItem, error)
}
