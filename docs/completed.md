# Completed Features

## Create Flow Improvements

**Problem**: Hardcoded Go constants for curated MCPs/skills — users couldn't customize. Create flow was linear, no way to reorder or remove selections.

**Solution**:
- Replaced hardcoded Go constants with YAML-backed curated config at `curated/mcps.yaml` + `curated/skills.yaml`, bundled via `embed.FS`
- User override support: `~/.config/aienv/curated/*.yaml` merged by name, tagged `(user override)` in menus
- Loop-based create flow — users freely mix curated selection, online search, and custom entries; `r` to remove; `d` to finish
- Deduplication: checks for existing name before appending
- Registry API parses `packages[].registryType` + `packages[].identifier` for proper command generation (npm→`npx -y`, pypi→`uvx`, go→`go run`)
- 12 of 20 curated MCPs include `env[]` metadata (key, description, required); env var requirements printed after selection
- Activation-time env var validation: `checkMCPEnvVars()` warns if any `env:KEY` reference has no matching shell variable
- Registry search results cross-referenced against curated entries to enrich with env var metadata

## Docker Sandbox Isolation

**Goal**: Run agents and MCP servers inside a Docker container, isolating them from the host.

**Approach**: Per-agent Dockerfiles (inline in Go source, Ubuntu 24.04 + Node.js 18 + Python 3.12 + Go 1.22), auto-built on first `--docker` use, `docker run --rm -it` with auto-cleanup on exit.

**Volume mounts**:
- `$(pwd):/workspace` (rw) — project source
- `~/.ai-envs/<name>/opencode.json:/ai-env/opencode.json:ro` — agent config
- `~/.config/opencode/skills:/home/user/.config/opencode/skills:ro` — installed skills
- `~/.config/opencode/:/home/user/.config/opencode:ro` — global config (providers, auth, themes)
- `~/.gitconfig:/home/user/.gitconfig:ro` — git authorship

**Bugs fixed during testing**:
- UID 1000 conflict with base image's `ubuntu` user → `userdel -r ubuntu` before `useradd`
- agent not in container → binary mount from host
- TUI not rendering → pass `TERM`/`COLORTERM` env vars
- eval subshell kills TTY → shell function bypasses `eval` when `--docker` is detected

## Starter Prompts

- `--prompt` flag on `aienv activate` to inject arbitrary starter text
- Optional `prompt` field in env schema for persisted defaults
- Prompt text written to `starter-prompt.md` and prepended to instructions array
- Runtime flag overrides env default
- Also added: self-referential `aienv` env with GitHub MCP + tdd/grill-me/caveman skills + AGENTS.md + Obsidian design notes as rules

## Multi-Agent Support

**Goal**: Support agents beyond OpenCode; clean pluggable architecture for per-agent config generation.

**Solution**:
- Agent interface: `internal/agents/agent.go` — `Agent` interface + global `Register()`/`Get()` registry
- OpenCode agent generates `opencode.json` with `mcp.<name>.command` as array
- Claude Code agent generates `mcp-config.json` with `mcpServers.<name>.command` (string) + `args` (array), `env` with `${VAR}` syntax. Remote MCPs skipped. Generates `CLAUDE.md` with prompt.
- Agent selection in create flow after name prompt, defaults to `"opencode"`
- Separate Dockerfiles per agent (`opencode.Dockerfile`, `claude.Dockerfile`) embedded via `//go:embed *.Dockerfile`
- `npx skills add` passes `--agent <agent>` for agent-scoped installs

## Config Inheritance + Docker Auth

**Problem**: Generated config was a flat struct with only `mcp`, `model`, `instructions`, `permission.skill` — all other user settings were dropped. Docker containers had no access to global config.

**Solution — minimal override generation**:
- Replaced rigid struct with `map[string]any` generation containing ONLY env-specific overrides
- OpenCode's native config merging handles inheritance of all other keys
- Global MCPs disabled at env level via `"enabled": false` for servers not in the env
- Permission deep-merge: `permission.skill` sub-key merged into global `permission` object
- Same code path for Docker and non-Docker

## Docker Sandbox Write Isolation

**Problem**: `~/.local/share/opencode/` was mounted as writable bind mount — agent writes leaked to host.

**Design evolution**: OverlayFS was the first approach but failed on Docker Desktop (kernel overlay not visible inside VM).

**Final approach**: Session-unique Docker named volume, initialized from host data, mounted writable. Writes go to `/var/lib/docker/volumes/`, never touch host.

**Implementation**:
- `docker run --rm --user root` to copy host data into volume, chown to uid 1000
- `sessionID = aienv-<envName>-<random>` per launch
- `defer` with `docker volume rm -f` on exit; `signal.Notify` for SIGINT/SIGTERM
- `~/.ssh/` removed — auth via `gh` CLI + token env vars
- `gh` CLI installed in both Dockerfiles

## Mount Isolated Volume Helper + Claude JSON Overlay

**Problem**: Claude Code inside Docker had `~/.claude/` mounted `:ro` (no write access to sessions, history). `~/.claude.json` not mounted — Claude started unauthenticated.

**Solution**:
- Extracted `mountIsolatedVolume(hostDir, volName, imageTag)` helper
- Refactored opencode `~/.local/share/opencode/` to use helper
- Changed claude `~/.claude/` from `:ro` to volume-init — writable, isolated from host
- Added `~/.claude.json` overlay: `:ro` mount + `sh -c "cp ...; exec claude ..."` wrapper

## Claude Code Config Inheritance

**Problem**: `CLAUDE_CONFIG_DIR=<envDir>/claude-config/` replaced `~/.claude/` entirely, dropping user's global config.

**Solution**:
- Removed `CLAUDE_CONFIG_DIR` export/unset — Claude uses global `~/.claude/`
- Env-specific overrides come via CLI flags (`--mcp-config`, `--append-system-prompt-file`, `--model`)
- Removed skill symlink logic — skills already installed globally by `npx skills add --agent claude`

## Default Environment Directory

**Problem**: Activation always stayed in the user's current directory. No way to configure a default workspace for an environment.

**Solution**:
- Added `workdir` field to `Env` struct (`yaml:"workdir,omitempty"`)
- Create flow prompts: "Default working directory (absolute path, or '.' for current): "
  - Empty → prints note about activation-time CWD behavior
  - `~` or `~/...` → expanded via `ExpandTilde()` helper
  - `.` or relative → converted to absolute via `filepath.Abs()`
  - Validates directory exists with `os.Stat`
- Activation (`cmd/activate_cmd.go`):
  - Resolves workdir, expands `~`, validates directory exists (warns + fallback to CWD if missing)
  - Passes resolved workdir to `GenerateFiles()` for correct rule path resolution
  - Non-Docker: prepends `cd <workdir>\n` to the activation command output → shell `eval` executes the `cd` in the calling shell
  - Docker: passes workdir to `docker.Run()` → mounts it as `/workspace` (container's `WORKDIR /workspace` inherited from Dockerfile)
- `show` and create summary display workdir when set
- `edit` works naturally — field is just another YAML key
- `ExpandTilde(path)` helper in `internal/env/env.go` shared across create and activate
