package opencode

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/kapilratnani/aienv/internal/agents"
	"github.com/kapilratnani/aienv/internal/env"
)

func init() {
	agents.Register(&agent{})
}

type agent struct{}

func (a *agent) Name() string { return "opencode" }

type mcpEntry struct {
	Type    string            `json:"type,omitempty"`
	Command []string          `json:"command,omitempty"`
	URL     string            `json:"url,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Enabled *bool             `json:"enabled,omitempty"`
}

func (a *agent) GenerateFiles(e *env.Env, cwd string) ([]agents.AgentFile, error) {
	globalCfg := readGlobalConfig()

	cfg := map[string]any{}
	cfg["$schema"] = "https://opencode.ai/config.json"

	if e.Model != "" {
		cfg["model"] = e.Model
	}

	if len(e.MCPServers) > 0 {
		cfg["mcp"] = buildMCP(e.MCPServers, globalCfg)
	}

	if len(e.Rules) > 0 {
		cfg["instructions"] = buildInstructions(e.Rules, cwd)
	}

	if perm := mergePermission(globalCfg, e.Skills, e.Permissions); len(perm) > 0 {
		cfg["permission"] = perm
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling opencode config: %w", err)
	}

	return []agents.AgentFile{
		{Path: "opencode.json", Content: data},
	}, nil
}

func (a *agent) ActivateCommand(envDir string, e *env.Env) string {
	absPath := filepath.Join(envDir, "opencode.json")

	shell := filepath.Base(os.Getenv("SHELL"))
	switch shell {
	case "fish":
		return fmt.Sprintf("set -x OPENCODE_CONFIG %s;\nopencode\nset -e OPENCODE_CONFIG;\n", absPath)
	default:
		return fmt.Sprintf("export OPENCODE_CONFIG=%s\nopencode\nunset OPENCODE_CONFIG\n", absPath)
	}
}

func readGlobalConfig() map[string]any {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	path := filepath.Join(home, ".config", "opencode", "opencode.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return cfg
}

func buildMCP(servers map[string]env.MCPServer, globalCfg map[string]any) map[string]mcpEntry {
	mcp := make(map[string]mcpEntry, len(servers))

	for name, srv := range servers {
		mcp[name] = mcpEntry{
			Type:    srv.Type,
			Command: srv.Command,
			URL:     srv.URL,
			Env:     srv.Env,
			Headers: srv.Headers,
		}
	}

	globalMCP, _ := globalCfg["mcp"].(map[string]any)
	for name := range globalMCP {
		if _, exists := servers[name]; !exists {
			f := false
			mcp[name] = mcpEntry{Enabled: &f}
		}
	}

	return mcp
}

func buildInstructions(rules []env.Rule, cwd string) []string {
	instructions := make([]string, 0, len(rules))
	for _, rule := range rules {
		path := rule.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(cwd, path)
		}
		instructions = append(instructions, path)
	}
	return instructions
}

func mergePermission(globalCfg map[string]any, skills []env.Skill, perms *env.Permissions) map[string]any {
	if len(skills) == 0 && perms == nil {
		return nil
	}

	perm, _ := globalCfg["permission"].(map[string]any)
	if perm == nil {
		perm = map[string]any{}
	} else {
		cp := make(map[string]any, len(perm))
		maps.Copy(cp, perm)
		perm = cp
	}

	skillPerm := map[string]string{"*": "deny"}
	for _, sk := range skills {
		skillPerm[sk.Name] = "allow"
	}
	perm["skill"] = skillPerm

	if perms != nil && perms.Filesystem != nil {
		if len(perms.Filesystem.Read) > 0 {
			existing, _ := perm["read"].(map[string]any)
			if existing == nil {
				merged := make(map[string]any, len(perms.Filesystem.Read))
				for k, v := range perms.Filesystem.Read {
					merged[k] = v
				}
				perm["read"] = merged
			} else {
				merged := make(map[string]any, len(existing)+len(perms.Filesystem.Read))
				maps.Copy(merged, existing)
				for k, v := range perms.Filesystem.Read {
					merged[k] = v
				}
				perm["read"] = merged
			}
		}
		if len(perms.Filesystem.Edit) > 0 {
			existing, _ := perm["edit"].(map[string]any)
			if existing == nil {
				merged := make(map[string]any, len(perms.Filesystem.Edit))
				for k, v := range perms.Filesystem.Edit {
					merged[k] = v
				}
				perm["edit"] = merged
			} else {
				merged := make(map[string]any, len(existing)+len(perms.Filesystem.Edit))
				maps.Copy(merged, existing)
				for k, v := range perms.Filesystem.Edit {
					merged[k] = v
				}
				perm["edit"] = merged
			}
		}
	}

	if perms != nil && len(perms.Bash) > 0 {
		existing, _ := perm["bash"].(map[string]any)
		if existing == nil {
			merged := make(map[string]any, len(perms.Bash))
			for k, v := range perms.Bash {
				merged[k] = v
			}
			perm["bash"] = merged
		} else {
			merged := make(map[string]any, len(existing)+len(perms.Bash))
			maps.Copy(merged, existing)
			for k, v := range perms.Bash {
				merged[k] = v
			}
			perm["bash"] = merged
		}
	}

	return perm
}
