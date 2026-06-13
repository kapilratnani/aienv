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

func (a *agent) DockerConfig(envDir string, e *env.Env, sessionID string) (*agents.DockerConfig, error) {
	home, _ := os.UserHomeDir()

	mounts := []string{
		fmt.Sprintf("%s:/workspace/.aienv:ro", envDir),
	}

	// Mount installed skill directories
	for _, s := range e.Skills {
		for _, prefix := range []string{".claude/skills"} {
			dir := filepath.Join(home, prefix, s.Name)
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				containerPath := fmt.Sprintf("/home/user/%s/%s", prefix, s.Name)
				mounts = append(mounts, fmt.Sprintf("%s:%s:ro", dir, containerPath))
				break
			}
		}
	}

	// Mount .claude.json as read-only, will be copied to writable path on startup
	claudeJSON := filepath.Join(home, ".claude.json")
	if _, err := os.Stat(claudeJSON); err == nil {
		mounts = append(mounts, fmt.Sprintf("%s:/home/user/.claude.json.ro:ro", claudeJSON))
	}

	claudeArgs := []string{
		"claude",
		"--mcp-config", "/workspace/.aienv/mcp-config.json",
		"--append-system-prompt-file", "/workspace/.aienv/CLAUDE.md",
		"--strict-mcp-config",
	}

	if e.Permissions != nil {
		claudeArgs = append(claudeArgs, "--settings", "/workspace/.aienv/claude-settings.json")
	}

	if e.Model != "" {
		claudeArgs = append(claudeArgs, "--model", e.Model)
	}

	for _, rule := range e.Rules {
		if !filepath.IsAbs(rule.Path) {
			claudeArgs = append(claudeArgs, "--append-system-prompt-file", filepath.Join("/workspace", rule.Path))
		} else {
			if _, err := os.Stat(rule.Path); err == nil {
				mounts = append(mounts, fmt.Sprintf("%s:%s:ro", rule.Path, rule.Path))
			}
			claudeArgs = append(claudeArgs, "--append-system-prompt-file", rule.Path)
		}
	}

	// Wrap with sh -c to copy .claude.json.ro → .claude.json on startup
	claudeCmd := strings.Join(claudeArgs, " ")
	entrypoint := []string{
		"aienv/sandbox:latest-claude",
		"sh", "-c",
		"cp /home/user/.claude.json.ro /home/user/.claude.json 2>/dev/null; exec " + claudeCmd,
	}

	return &agents.DockerConfig{
		Mounts:     mounts,
		Entrypoint: entrypoint,
	}, nil
}
