# Docker-Only Migration

Last updated: 2026-06-13
Status: Complete

## Quick Resume

Remove all non-Docker code paths. aienv becomes a Docker-only sandbox for running terminal-based AI agents. All activation launches a container. No shell function, no eval, no `--docker` flag — just `docker run`.

## Decisions

| # | Decision |
|---|---|
| Agent scope | Docker-only for all terminal agents (OpenCode, Claude Code first; Codex, Copilot Dockerfiles follow-up) |
| Shell function | **Drop** — no eval, no subshell, no shell wrapper |
| Workdir | **Required** at create time. Mounted as `/workspace`. No CWD fallback. |
| Dead packages | **Delete** `internal/shell/`, `cmd/init_cmd.go` |
| `aienv up` | **Keep** — discovers `.aienv.yaml`, registers, calls `docker run` directly (no eval) |
| `--docker` flag | **Remove** — every activate is Docker |
| Docker as hard dep | **Accept** — add post-install check that warns if `docker` not found |
| `--model` / `--prompt` | Forwarded into container as before — no change |
| Codex/Copilot support | Follow-up PRs once install paths confirmed |

## Agent Interface Refactor

### Problem

- `ActivateCommand()` returns a shell string — only used by non-Docker mode (dead code after migration)
- `docker/sandbox.go` has a giant `switch e.Agent` block with hardcoded mounts, env vars, and entrypoint per agent — violates Open/Closed Principle

### Solution

Replace `ActivateCommand()` with `DockerConfig()`:

```go
// internal/agents/agent.go
type DockerConfig struct {
    Mounts     []string // "-v" arguments (config files, skills, etc.)
    EnvVars    []string // "-e" arguments
    Entrypoint []string // ["image:tag", "command", "--flag", ...]
}

type Agent interface {
    Name() string
    GenerateFiles(e *env.Env, cwd string) ([]AgentFile, error)
    DockerConfig(envDir string, e *env.Env, sessionID string) (*DockerConfig, error)
}
```

Each agent implements `DockerConfig()` with its own mount/entrypoint logic. The `sandbox.go` switch statement disappears — it calls `agent.DockerConfig()` and appends generic mounts (workdir, gitconfig, TTY env vars, etc.).

### OpenCode `DockerConfig` Example

```go
func (a *agent) DockerConfig(envDir, sessionID string, e *env.Env) (*DockerConfig, error) {
    cfg := &DockerConfig{
        EnvVars:    []string{"OPENCODE_CONFIG=/ai-env/opencode.json"},
        Entrypoint: []string{imageTag("opencode"), "opencode"},
    }
    cfg.Mounts = append(cfg.Mounts,
        fmt.Sprintf("%s/opencode.json:/ai-env/opencode.json:ro", envDir))
    // skill mounts...
    return cfg, nil
}
```

### Claude Code `DockerConfig` Example

```go
func (a *agent) DockerConfig(envDir, sessionID string, e *env.Env) (*DockerConfig, error) {
    cfg := &DockerConfig{}
    cfg.Mounts = append(cfg.Mounts,
        fmt.Sprintf("%s/mcp-config.json:/ai-env/mcp-config.json:ro", envDir),
        fmt.Sprintf("%s/CLAUDE.md:/ai-env/CLAUDE.md:ro", envDir))
    if e.Permissions != nil {
        cfg.Mounts = append(cfg.Mounts,
            fmt.Sprintf("%s/claude-settings.json:/ai-env/claude-settings.json:ro", envDir))
    }

    claudeArgs := []string{"--mcp-config", "/ai-env/mcp-config.json",
        "--append-system-prompt-file", "/ai-env/CLAUDE.md", "--strict-mcp-config"}
    if e.Model != "" {
        claudeArgs = append(claudeArgs, "--model", e.Model)
    }
    // .claude.json overlay wrapper
    cfg.Entrypoint = []string{imageTag("claude-code"), "sh", "-c",
        fmt.Sprintf("cp /home/user/.claude.json.ro /home/user/.claude.json 2>/dev/null; exec claude %s",
            strings.Join(claudeArgs, " "))}
    return cfg, nil
}
```

## Update Existing Docs (After Implementation)

The following 8 files reference non-Docker mode and must be updated once the code migration is complete:

| File | Issues | Lines |
|---|---|---|
| `architecture.md` | Lists `init` command, shell function `eval` description, `ActivateCommand()` in interface | 12, 32, 34, 41, 45 |
| `roadmap.md` | Describes non-Docker workdir prepend as current | 48-49 |
| `completed.md` | `--docker` flag refs, eval subshell fix, non-Docker workdir, "same code path" | 21, 34, 65, 115-116 |
| `custom-docker-images.md` | Activation example uses `--docker` flag | 106 |
| `docker.md` | "Auto-built on first `--docker` use" | 45 |
| `desktop-ui.md` | Dual-mode terminal handling, non-Docker PTY path, `ActivateCommand()` ref | 71, 95-98, 110-111, 138, 218, 289, 292, 305-306 |
| `audit-logs.md` | Open question about non-Docker cost extraction | 79 |
| `trust-permissions.md` | References `ActivateCommand()` for claude settings | 47 |

### `sandbox.go` `Run()` Becomes

```go
func Run(e *env.Env, workdir string) error {
    if err := Check(); err != nil { return err }
    if err := ensureImage(e.Agent); err != nil { return err }

    sessionID := fmt.Sprintf("aienv-%s-%d", e.Name, rand.Uint64())
    // signal handling, volume cleanup...

    ag, _ := agents.Get(e.Agent)
    cfg, _ := ag.DockerConfig(config.EnvDir(e.Name), sessionID, e)

    args := []string{"run", "--rm", "-it"}
    args = append(args, "-v", fmt.Sprintf("%s:/workspace", workdir))
    // generic mounts: gitconfig, global config dirs, isolated volumes...
    args = append(args, cfg.Mounts...)
    // TTY env vars + MCP env vars...
    args = append(args, cfg.EnvVars...)
    // proxy setup if permissions.network...
    args = append(args, cfg.Entrypoint...)

    cmd := exec.Command("docker", args...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

## Changes by File

| File | Change |
|---|---|
| `internal/agents/agent.go` | Remove `ActivateCommand()`, add `DockerConfig()` method + `DockerConfig` struct |
| `internal/agents/opencode/gen.go` | Remove `ActivateCommand()`, implement `DockerConfig()` |
| `internal/agents/claude/gen.go` | Remove `ActivateCommand()`, implement `DockerConfig()` |
| `internal/docker/sandbox.go` | Remove `switch e.Agent` blocks, call `agent.DockerConfig()` instead |
| `cmd/activate_cmd.go` | Remove non-Docker path, remove `dockerMode` check, always call `docker.Run()` |
| `cmd/root.go` | Remove `dockerMode` flag var, remove `--docker` persistent flag, update help text |
| `cmd/init_cmd.go` | **Delete** entire file |
| `internal/shell/init.go` | **Delete** entire package |
| `internal/env/env.go` | `workdir` becomes required — validate non-empty in `Save()` |
| `cmd/create_cmd.go` | Remove optional workdir note (lines 649-653), require workdir |

## Files That Stay the Same

- `internal/env/yaml.go` — serialization unchanged
- `internal/docker/embed.go` — still reads embedded Dockerfiles
- `internal/docker/proxy.go` — already Docker-only
- `internal/registry/` — MCP/skill search unchanged
- `internal/skills/` — skill install unchanged
- `internal/assets/` — curated config loader unchanged
- `internal/config/paths.go` — path helpers unchanged
- `curated/` — YAML files unchanged
- `cmd/permissions_cmd.go` — unchanged (uses agent interface)
- `cmd/show_cmd.go`, `cmd/edit_cmd.go`, `cmd/list_cmd.go`, `cmd/delete_cmd.go` — unchanged
- `agent_import.go` — unchanged (agent registrations stay)
