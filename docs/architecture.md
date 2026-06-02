# Architecture

## Tech Stack

- **Language**: Go 1.26.2, module `github.com/kapilratnani/aienv`
- **CLI Framework**: `spf13/cobra`
- **YAML**: `gopkg.in/yaml.v3`

## Project Layout

```
cmd/              — Thin Cobra commands (create, activate, list, show, edit, delete, init, docker)
internal/         — Core logic (unexported packages)
  env/            — Env struct, YAML load/save/validate
  agents/         — Agent interface + per-agent config generation
    opencode/     — OpenCode agent (opencode.json generation)
    claude/       — Claude Code agent (mcp-config.json + CLAUDE.md generation)
  skills/         — Skill verify and install
  registry/       — MCP registry and skills.sh API clients
  assets/         — Curated MCP/skill YAML loader (embed.FS + user overrides)
  docker/         — Docker sandbox container (build + run)
  shell/          — Shell function install
  config/         — Path helpers
curated/          — YAML-backed curated MCP and skill lists (embedded via embed.FS)
examples/         — Reference env YAML files
docs/             — Documentation
```

## Activation Model

- Environments stored at `~/.ai-envs/<name>/`
- Shell function (`aienv()`) wraps the binary with `eval "$(aienv activate "$@")"` — foreground blocking
- `OPENCODE_CONFIG` is exported to point to generated config, unset on exit
- Docker mode bypasses `eval` entirely to preserve real TTY — calls `command aienv activate "$@"` directly

## Config Inheritance

- Generated config contains **only** env-specific overrides (`mcp`, `model`, `instructions`, `permission`)
- OpenCode's native config merging (global → `OPENCODE_CONFIG`) handles inheritance of all other keys
- Global MCPs are disabled at the env level by emitting `"enabled": false` for servers not in the env
- Same code path for Docker and non-Docker — no generation-time deep-copy

## Multi-Agent Architecture

- `internal/agents/agent.go` defines the `Agent` interface (`Name()`, `GenerateFiles()`, `ActivateCommand()`)
- Global `Register()` / `Get()` registry for pluggable agents
- New agents register via blank import in `agent_import.go`
- Supported: OpenCode, Claude Code

## Curated Config

- YAML files at `curated/mcps.yaml` + `curated/skills.yaml`, bundled via `embed.FS`
- User overrides: `~/.config/aienv/curated/*.yaml` merged by name, tagged `(user override)` in menus
- Registry API parses `packages[].registryType` + `packages[].identifier` for command generation:
  - npm → `npx -y`
  - pypi → `uvx`
  - go → `go run`

## Schema

- Flat YAML, no anchors, no plugins/context fields
- Fields: `name`, `agent`, `model`, `prompt`, `mcp`, `skills`, `rules`

## Market Research

- MCP ecosystem: 10K+ servers available, 97M+ SDK downloads
- #1 pain point: context token bloat from too many active MCPs
- aienv positioned as "virtualenv for AI" — encapsulates per-task MCPs, skills, and rules
