# Architecture

## Tech Stack

- **Language**: Go 1.26.2, module `github.com/kapilratnani/aienv`
- **CLI Framework**: `spf13/cobra`
- **YAML**: `gopkg.in/yaml.v3`

## Project Layout

```
cmd/              — Thin Cobra commands (create, activate, list, show, edit, delete, build, clean)
internal/         — Core logic (unexported packages)
  audit/          — JSONL audit log schema and writer
  config/         — XDG paths, hash computation, session ID generation, trust cache paths
  docker/         — Docker sandbox (build, run, proxy, trust prompts)
  env/            — Env struct, YAML load/save/validate
docs/             — Documentation, ADRs, agent skill references
```

## Activation Model

- Environments stored at `~/.local/share/aienv/<name>/env.yaml` (XDG-compliant)
- Each `aienv up` generates a session ID, creates an audit dir, and runs a new `--read-only` container
- Image cached by env.yaml content hash; auto-rebuilt on content change
- Network proxy runs on the host and enforces allow/deny/learn rules
- Audit logs written to `~/.local/share/aienv/<name>/audit/<session-id>/`

## Trust System

- First activation shows a trust prompt with mounts and network rules
- Trust cached at `~/.config/aienv/trust/<content-hash>.json`
- Cache invalidated when env.yaml changes (different hash)
- `aienv clean` clears orphaned trust cache entries

## Audit Logging

- Session metadata written as `session.meta.json` on the host
- Network requests logged as JSONL to `network.jsonl` in the audit dir
- Writer runs on the host (proxy process); container sees audit dir as a writable bind mount

## Black Box Agent Architecture

- Agents are specified entirely through YAML (install, command, mounts, env vars)
- aienv never generates agent config files, installs MCPs, or understands agent-specific formats
- Per-agent Go implementations were deleted in favor of this generic approach

## Schema

```yaml
env:
  name: my-env
  description: My AI environment
agent:
  install:
    - npm install -g opencode-ai
  command: [opencode]
  prompt_flag: "-p"          # optional: flag for aienv up --prompt/-p to send prompt to agent
  mounts:
    - source: ~/project
      target: /workspace
    - source: ~/.config/opencode
      target: /home/agent/.config/opencode
      writable: true
deps:
  packages: [golang, nodejs]
  custom: [go install foo/bar]
permissions:
  network:
    allow: [api.github.com, api.anthropic.com]
    deny: [*]
    learn: false
audit:
  persist: true
  capture: [network]
```
