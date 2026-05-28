package opencode

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/kapilratnani/aienv/internal/env"
)

type opencodeConfig struct {
	Model        string              `json:"model,omitempty"`
	MCP          map[string]mcpEntry `json:"mcp"`
	Instructions []string            `json:"instructions"`
	Permission   *permissionConfig   `json:"permission,omitempty"`
}

type permissionConfig struct {
	Skill map[string]string `json:"skill"`
}

type mcpEntry struct {
	Type    string            `json:"type,omitempty"`
	Command []string          `json:"command,omitempty"`
	URL     string            `json:"url,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

func Generate(e *env.Env, cwd string) ([]byte, error) {
	cfg := opencodeConfig{
		MCP:          make(map[string]mcpEntry, len(e.MCPServers)),
		Instructions: make([]string, 0, len(e.Rules)),
	}

	if e.Model != "" {
		cfg.Model = e.Model
	}

	for name, srv := range e.MCPServers {
		cfg.MCP[name] = mcpEntry{
			Type:    srv.Type,
			Command: srv.Command,
			URL:     srv.URL,
			Env:     srv.Env,
			Headers: srv.Headers,
		}
	}

	for _, rule := range e.Rules {
		path := rule.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(cwd, path)
		}
		cfg.Instructions = append(cfg.Instructions, path)
	}

	skillPerm := map[string]string{"*": "deny"}
	for _, sk := range e.Skills {
		skillPerm[sk.Name] = "allow"
	}
	cfg.Permission = &permissionConfig{Skill: skillPerm}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling opencode config: %w", err)
	}

	return data, nil
}
