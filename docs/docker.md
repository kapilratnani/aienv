# Docker Sandbox

## Overview

The Docker sandbox runs AI coding agents inside isolated containers, preventing agent writes from leaking to the host filesystem while preserving access to project source code and authentication.

## Volume Mounts

| Host Path | Container Path | Mode | Purpose |
|---|---|---|---|
| `$(pwd)` | `/workspace` | rw | Project source code |
| `~/.ai-envs/<name>/` config | `/ai-env/` | ro | Agent config files |
| `~/.config/opencode/` | `/home/user/.config/opencode/` | ro | Global config, providers, auth |
| `~/.gitconfig` | `/home/user/.gitconfig` | ro | Git commit authorship |

## Isolated Volumes (Session-Unique)

Uses Docker named volumes initialized from host data — writes go to `/var/lib/docker/volumes/`, never touch host.

| Host Path | Volume Name | Container Path | Strategy |
|---|---|---|---|
| `~/.local/share/opencode/` | `aienv-<name>-<random>` | `/home/user/.local/share/opencode/` | volume-init |
| `~/.claude/` | `aienv-claude-<name>-<random>` | `/home/user/.claude/` | volume-init |
| `~/.claude.json` | — | `/home/user/.claude.json.ro` (ro) + cp to writable | overlay-init |

### volume-init

```go
func mountIsolatedVolume(hostDir, volName, imageTag string) (string, error)
```

1. Create a Docker named volume
2. Run ephemeral root container: `cp -a /source/. /target/. && chown -R 1000:1000 /target`
3. Mount the volume at the target path

### overlay-init

For `~/.claude.json`: mount `:ro` at `.claude.json.ro`, then wrap the entrypoint with `sh -c "cp /home/user/.claude.json.ro /home/user/.claude.json && exec <agent> ..."`.

## Lifecycle

- **Container**: `docker run --rm -it` — auto-cleanup on exit
- **Volumes**: `defer docker volume rm -f` on normal exit
- **Signals**: `signal.Notify` catches SIGINT/SIGTERM before Go exits, ensuring defer runs
- **Image**: Auto-built on first `--docker` use, rebuild via `aienv docker build`

## Per-Agent Dockerfiles

Embedded via `//go:embed *.Dockerfile` in `internal/docker/`:
- `opencode.Dockerfile` — installs opencode via npm, includes gh CLI
- `claude.Dockerfile` — installs claude-code via npm, includes gh CLI

Image tag: `aienv/sandbox:latest-<agent>`

## Network Policy Proxy

When `permissions.network` is configured, an embedded Go HTTP/HTTPS proxy enforces the domain allowlist/denylist:

1. Proxy binds to `0.0.0.0` (all host interfaces) on a random port
2. Container resolves the host via `--add-host host.docker.internal:host-gateway` (`host.docker.internal`)
3. Container receives `HTTP_PROXY`/`HTTPS_PROXY` pointing to `host.docker.internal:<port>`
4. Container uses default Docker bridge network (no `--network host`)

## Environment Passthrough

`TERM`, `COLORTERM`, `LANG`, `LC_ALL`, `LC_CTYPE`, `OPENCODE_CONFIG`, `HOME`, MCP-specific vars (`GITHUB_TOKEN`, etc.)

## Known Issues

- `~/.ssh/` not mounted — auth via `gh` CLI + token env vars
