package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kapilratnani/aienv/internal/agents"
	"github.com/kapilratnani/aienv/internal/env"
)

func init() {
	agents.Register(&agent{})
}

type agent struct{}

func (a *agent) Name() string { return "claude-code" }

type claudeConfig struct {
	MCPServers map[string]mcpEntry `json:"mcpServers"`
}

type claudeSettings struct {
	Permissions *claudePermissions `json:"permissions,omitempty"`
}

type claudePermissions struct {
	Allow []string `json:"allow,omitempty"`
	Ask   []string `json:"ask,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

type mcpEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func (a *agent) GenerateFiles(e *env.Env, cwd string) ([]agents.AgentFile, error) {
	var files []agents.AgentFile

	cfg := claudeConfig{
		MCPServers: make(map[string]mcpEntry, len(e.MCPServers)),
	}

	for name, srv := range e.MCPServers {
		if srv.Type == "remote" {
			fmt.Fprintf(os.Stderr, "  Warning: skipping remote MCP %q — Claude Code --mcp-config only supports local servers\n", name)
			continue
		}
		entry := mcpEntry{}
		if len(srv.Command) > 0 {
			entry.Command = srv.Command[0]
			if len(srv.Command) > 1 {
				entry.Args = srv.Command[1:]
			}
		}
		if len(srv.Env) > 0 {
			entry.Env = make(map[string]string, len(srv.Env))
			for k, v := range srv.Env {
				if len(v) > 4 && v[:4] == "env:" {
					entry.Env[k] = fmt.Sprintf("${%s}", v[4:])
				} else {
					entry.Env[k] = v
				}
			}
		}
		cfg.MCPServers[name] = entry
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling claude mcp config: %w", err)
	}
	files = append(files, agents.AgentFile{Path: "mcp-config.json", Content: data})

	var promptLines []string

	if e.Prompt != "" {
		promptLines = append(promptLines, e.Prompt)
	}

	if len(promptLines) > 0 {
		claudeMD := strings.Join(promptLines, "\n") + "\n"
		files = append(files, agents.AgentFile{
			Path:    "CLAUDE.md",
			Content: []byte(claudeMD),
		})
	} else {
		files = append(files, agents.AgentFile{
			Path:    "CLAUDE.md",
			Content: []byte{},
		})
	}

	if e.Permissions != nil {
		settingsData, err := json.MarshalIndent(buildClaudeSettings(e.Permissions), "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshalling claude settings: %w", err)
		}
		files = append(files, agents.AgentFile{
			Path:    "claude-settings.json",
			Content: settingsData,
		})
	}

	return files, nil
}

func buildClaudeSettings(perms *env.Permissions) claudeSettings {
	if perms == nil {
		return claudeSettings{}
	}
	cp := claudePermissions{}
	addTo := func(action string, entry string) {
		switch action {
		case "allow":
			cp.Allow = append(cp.Allow, entry)
		case "ask":
			cp.Ask = append(cp.Ask, entry)
		case "deny":
			cp.Deny = append(cp.Deny, entry)
		}
	}
	if perms.Filesystem != nil {
		for pattern, action := range perms.Filesystem.Read {
			addTo(action, "Read("+pattern+")")
		}
		for pattern, action := range perms.Filesystem.Edit {
			addTo(action, "Edit("+pattern+")")
		}
	}
	for pattern, action := range perms.Bash {
		addTo(action, "Bash("+pattern+")")
	}
	settings := claudeSettings{}
	if len(cp.Allow) > 0 || len(cp.Ask) > 0 || len(cp.Deny) > 0 {
		settings.Permissions = &cp
	}
	return settings
}

func (a *agent) ActivateCommand(envDir string, e *env.Env) string {
	var args []string
	args = append(args, "claude")
	args = append(args, "--mcp-config", filepath.Join(envDir, "mcp-config.json"))
	args = append(args, "--append-system-prompt-file", filepath.Join(envDir, "CLAUDE.md"))

	if e.Permissions != nil {
		args = append(args, "--settings", filepath.Join(envDir, "claude-settings.json"))
	}

	for _, rule := range e.Rules {
		path := rule.Path
		if !filepath.IsAbs(path) {
			cwd, _ := os.Getwd()
			path = filepath.Join(cwd, path)
		}
		args = append(args, "--append-system-prompt-file", path)
	}

	args = append(args, "--strict-mcp-config")

	if e.Model != "" {
		args = append(args, "--model", e.Model)
	}

	return strings.Join(args, " ") + "\n"
}
