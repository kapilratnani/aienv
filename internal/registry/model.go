package registry

import "context"

type MCPServerItem struct {
	Name        string `json:"name"`
	Command     string `json:"command,omitempty"`
	Type        string `json:"type"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

type SkillItem struct {
	Name        string `json:"name"`
	Package     string `json:"package"`
	Description string `json:"description,omitempty"`
	Installs    int    `json:"installs,omitempty"`
}

type MCPRegistry interface {
	Search(ctx context.Context, query string, limit int) ([]MCPServerItem, error)
}

type SkillRegistry interface {
	Search(ctx context.Context, query string, limit int) ([]SkillItem, error)
}
