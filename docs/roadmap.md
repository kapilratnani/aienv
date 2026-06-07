# Roadmap

## Priority 1: Desktop UI

- Wails-based desktop app at `cmd/wails/` — separate binary, reuses `internal/` packages
- Vue 3 + TypeScript + Vite frontend with xterm.js embedded terminal
- Full replacement of CLI for common tasks: env CRUD, activation, permissions, audit
- Phase 1: Shell & Env CRUD (dashboard, create wizard, detail/edit views)
- Phase 2: Permissions & Trust (visual glob editor, provider endpoints, OS keychain)
- Phase 3: Embedded Terminal (Terminal interface, creack/pty, xterm.js, Docker session launch)
- Phase 4: Audit Viewer (network JSONL, cost extraction, session list/detail)
- Separate `go.mod` for Wails — no dependency bloat on CLI binary
- Linux + macOS only (Windows deferred to ConPTY)
- Design doc: `docs/desktop-ui.md` — 26 grill-session decisions logged

## Priority 2: Repo-Local `.aienv.yaml` + `aienv up`

- `.aienv.yaml` as canonical repo-level env declaration (checked into VCS)
- Discovery algorithm: walk up from CWD to git root
- Registration symlink at `~/.ai-envs/_repo_/<dirname>/`
- Shell function: `aienv up` reserved subcommand
- `aienv down` to unregister repo-local env
- `.aienv.lock` for reproducible environments (pins MCP/skill versions)
- `--frozen` mode for CI — fails if lockfile would change
- `aienv init-repo` to scaffold `.aienv.yaml` from project type detection

## Priority 3: Permission Policies & Trust

- `permissions` struct in schema: `filesystem` (writable/readonly), `network` (allow/deny), `bash` (allow/deny)
- Permission to OpenCode `opencode.json` permission config translation
- Permission to Docker sandbox enforcement (read-only mounts, network rules)
- Interactive trust/review prompt on first `aienv up` for unknown repos
- Trust cache: `~/.config/aienv/trust/<content-hash>.json`, invalidated on `.aienv.yaml` changes
- Future: permission signing via GPG/Sigstore for trustless verification

## Priority 4: Agent Expansion Framework

- Redesign agent architecture for simplicity — most agents share common patterns (MCPs as JSON, instructions as markdown rules files)
- Extract a base/default agent with overridable paths, file templates, and activation command patterns
- Deferred: actual per-agent implementations (Cursor, GitHub Copilot, Windsurf, Codex) until framework is stable

## Priority 5: Default Environment Directory (DONE)

- On activation, change directory to a configured workspace path
- `workdir` field in the env schema (absolute path, stored in YAML)
- Create flow prompts for workdir with tilde/relative path expansion and directory validation
- Activation resolves workdir, passes it to GenerateFiles for rule path resolution
- Non-Docker: prepends `cd <workdir>` to activation command (eval picks it up)
- Docker: mounts workdir as `/workspace` (instead of CWD)
- `show` and create summary display the workdir setting
- `ExpandTilde()` helper in `internal/env/env.go` for `~` expansion at both create and activation time

## Priority 6: Custom MCP/Skill Repositories

- Support additional registries beyond skills.sh and modelcontextprotocol.io
- Enterprise/internal repo support via configurable registry list
- Multi-registry orchestration (merge results from all configured repos)

## Priority 7: Sharing & Team Features

- `aienv install <source>` — install environments from GitHub repos or URLs
- `aienv publish` — export environment to GitHub
- `aienv update` — pull latest version of a shared environment
- GitHub-based discovery: search for `.aienv.yaml` files

## Priority 8: Docker Image Size

- Multi-stage Dockerfiles, distroless base, smaller image
- Not a blocker for local development
